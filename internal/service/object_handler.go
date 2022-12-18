package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	tools "bb-project/tool"
)

type ObjectDataPort interface {
	SaveObjects(context.Context, []Object) error
}

type ObjectHandler struct {
	client   *http.Client
	data     ObjectDataPort
	endpoint string
	wg       *sync.WaitGroup
}

func NewObjectHandler(dataPort ObjectDataPort, endpoint string) *ObjectHandler {
	// Customize the Transport to have larger connection pool
	transport := http.DefaultTransport.(*http.Transport)
	transport.MaxIdleConns = 1000
	transport.MaxIdleConnsPerHost = 1000
	return &ObjectHandler{
		client:   &http.Client{Transport: transport},
		data:     dataPort,
		endpoint: endpoint,
		wg:       &sync.WaitGroup{},
	}
}

// Handle Perform batching object processing
// Each id will be processed concurrently
func (s *ObjectHandler) Handle(ctx context.Context, msg []string) error {
	ids := reduce(parse(msg))
	objList := make([]Object, len(ids))

	for k, id := range ids {
		s.wg.Add(1)
		objList[k] = idToObject(id)
		go func(wg *sync.WaitGroup, object *Object) {
			defer wg.Done()
			err := s.httpHandler(ctx, object)
			if err != nil {
				log.Err(err).Send()
			}
			object.LastSeen = time.Now().UTC()
		}(s.wg, &objList[k])
	}
	s.wg.Wait()
	if ctx.Err() != nil {
		// Handle context cancellation for exit
		log.Debug().Msg("Context canceled, handler stopped")
		return ctx.Err()
	}
	err := s.saveObjects(ctx, objList)
	if err != nil {
		log.Err(err).Send()
		return err
	}
	return nil
}

// httpHandler calls the object endpoint until success or the context cancellation
func (s *ObjectHandler) httpHandler(ctx context.Context, object *Object) error {
	for ctx.Err() == nil {
		err := s.do(ctx, object)
		if err == nil {
			return nil
		}
		if errors.Is(err, context.Canceled) {
			log.Info().Msg(err.Error())
			return nil
		}
		log.Err(err).Msg("httpHandler request error, retry...")
		tools.Sleep(ctx, 1*time.Second)
	}
	return ctx.Err()
}

func (s *ObjectHandler) do(ctx context.Context, object *Object) error {
	url := s.endpoint + strconv.Itoa(object.Id)
	ctx1, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx1, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request finished with code %d", resp.StatusCode)
	}
	err = json.NewDecoder(resp.Body).Decode(object)
	if err != nil {
		return fmt.Errorf("request finished with code %d: %w", resp.StatusCode, err)
	}
	return nil
}

// saveObjects calls the SaveObjects until success or the context cancellation
func (s *ObjectHandler) saveObjects(ctx context.Context, objectList []Object) error {
	objectList = filteredOut(objectList)
	if len(objectList) == 0 {
		return nil
	}
	for ctx.Err() == nil {
		err := s.data.SaveObjects(ctx, objectList)
		if err == nil {
			return nil
		}
		if errors.Is(err, context.Canceled) {
			return err
		}
		log.Err(err).Msg("saveObjects request error, retry...")
		tools.Sleep(ctx, 1*time.Second)
	}
	return ctx.Err()
}

func filteredOut(obj []Object) (res []Object) {
	if len(obj) == 0 {
		return
	}
	for k := range obj {
		if obj[k].Online {
			res = append(res, obj[k])
		}
	}
	return
}

func parse(msgs []string) []int {
	res := make([]int, 0, len(msgs)*8)
	for _, msg := range msgs {
		var ids []int
		err := json.Unmarshal([]byte(msg), &ids)
		if err != nil {
			log.Err(err).Msg("wrong data")
			continue
		}
		if len(ids) > 0 {
			res = append(res, ids...)
		}
	}
	return res
}

func reduce(intList []int) []int {
	allKeys := make(map[int]bool)
	list := []int{}
	for _, item := range intList {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}
