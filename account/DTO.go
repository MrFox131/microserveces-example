package account

import (
	"context"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/streadway/amqp"
	"net/http"
)

type (
	CreateUserRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	CreateUserResponse struct {
		Ok string `json:"ok"`
	}

	GetUserRequest struct {
		ID string `json:"id"`
	}
	GetUserResponse struct {
		Email string `json:"email"`
	}
)

func encodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	return json.NewEncoder(w).Encode(response)
}

func decodeUserReq(_ context.Context, r *http.Request) (interface{}, error) {
	var req CreateUserRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func decodeEmailReq(_ context.Context, r *http.Request) (interface{}, error) {
	var req GetUserRequest
	vars := mux.Vars(r)

	req = GetUserRequest{
		ID: vars["ID"],
	}

	return req, nil
}

func decodeUserAMQPReq(_ context.Context, delivery *amqp.Delivery) (request interface{}, err error) {
	var req CreateUserRequest
	if err := json.Unmarshal(delivery.Body, &req); err != nil {
		return nil, err
	}

	return req, nil
}

func decodeEmailAMQPReq(_ context.Context, delivery *amqp.Delivery) (request interface{}, err error) {
	var req GetUserRequest
	if err := json.Unmarshal(delivery.Body, &req); err != nil {
		return nil, err
	}

	return req, nil
}

func encodeAMQPResponse(_ context.Context, amqpPublishing *amqp.Publishing, res interface{}) error {
	responseString, err := json.Marshal(res)
	if err != nil {
		return err
	}

	amqpPublishing.Body = responseString
	return nil
}
