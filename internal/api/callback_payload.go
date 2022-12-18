package api

type CallbackRequest struct {
	ObjectIds []int `json:"object_ids"`
}

type ObjectResponse struct {
	Id     int  `json:"id"`
	Online bool `json:"online"`
}
