package service

import (
	"context"
	"io"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/golang/mock/gomock"
	c "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
)


type MockRoundTripper func(r *http.Request) *http.Response

func (f MockRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r), nil
}

func TestObjectHandler_do(t *testing.T) {
	c.Convey("do", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		a := assert.New(t)

		mData := NewMockObjectDataPort(ctrl)
		service := NewObjectHandler(mData, "http://localhost:9010/objects/")

		c.Convey("No errors", func() {
			service.client = &http.Client{
				Transport: MockRoundTripper(func(r *http.Request) *http.Response {
					return &http.Response{
						StatusCode: 200,
						Body:       io.NopCloser(strings.NewReader(`{"id":23,"online":true}`)),
					}
				}),
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			object := &Object{Id: 23}
			objectExp := &Object{Id: 23, Online: true}

			err := service.do(ctx, object)

			a.NoError(err)
			a.Equal(objectExp, object)
		})

		c.Convey("Not found", func() {
			service.client = &http.Client{
				Transport: MockRoundTripper(func(r *http.Request) *http.Response {
					return &http.Response{
						StatusCode: 404,
					}
				}),
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			object := &Object{Id: 23}
			objectExp := &Object{Id: 23}

			err := service.do(ctx, object)

			a.EqualError(err, "request finished with code 404")
			a.Equal(objectExp, object)
		})
	})
}

func TestObjectHandler_httpHandler(t *testing.T) {
	type fields struct {
		client   *http.Client
		data     ObjectDataPort
		endpoint string
		wg       *sync.WaitGroup
	}
	type args struct {
		ctx    context.Context
		object *Object
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ObjectHandler{
				client:   tt.fields.client,
				data:     tt.fields.data,
				endpoint: tt.fields.endpoint,
				wg:       tt.fields.wg,
			}
			if err := s.httpHandler(tt.args.ctx, tt.args.object); (err != nil) != tt.wantErr {
				t.Errorf("httpHandler() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestObjectHandler_saveObjects(t *testing.T) {
	c.Convey("saveObjects", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		a := assert.New(t)

		mData := NewMockObjectDataPort(ctrl)
		service := NewObjectHandler(mData, "http://localhost:9010/objects/")

		oblList := []Object{{Id: 12, Online: true}}

		c.Convey("Success", func() {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			mData.EXPECT().SaveObjects(ctx, oblList).Return(nil)
			err := service.saveObjects(ctx, oblList)

			a.NoError(err)
		})

		c.Convey("Context Canceled", func() {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			err := service.saveObjects(ctx, oblList)

			a.ErrorIs(err, context.Canceled)
		})
	})
}

func Test_filteredOut(t *testing.T) {
	type args struct {
		obj []Object
	}
	tests := []struct {
		name    string
		args    args
		wantRes []Object
	}{
		{
			name: "success",
			args: args{[]Object{
				{Id: 12, Online: true},
				{Id: 15, Online: false},
				{Id: 33, Online: false},
				{Id: 78, Online: true},
			}},
			wantRes: []Object{
				{Id: 12, Online: true},
				{Id: 78, Online: true},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotRes := filteredOut(tt.args.obj); !reflect.DeepEqual(gotRes, tt.wantRes) {
				t.Errorf("filteredOut() = %v, want %v", gotRes, tt.wantRes)
			}
		})
	}
}

func Test_parse(t *testing.T) {
	type args struct {
		msgs []string
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		{
			name: "success",
			args: args{msgs: []string{
				"[2,98,12,67]",
				"[33,98,84,0,5]",
				"[]",
			}},
			want: []int{2, 98, 12, 67, 33, 98, 84, 0, 5},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parse(tt.args.msgs); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_reduce(t *testing.T) {
	a := assert.New(t)
	type args struct {
		intList []int
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		{
			name: "success",
			args: args{intList: []int{2, 98, 12, 67, 33, 98, 84, 0, 5}},
			want: []int{2, 98, 12, 67, 33, 84, 0, 5},
		},
		{
			name: "success empty",
			args: args{intList: []int{}},
			want: []int{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := reduce(tt.args.intList)
			a.Equal(tt.want, got)
		})
	}
}
