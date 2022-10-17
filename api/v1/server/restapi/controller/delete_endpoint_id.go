// Code generated by go-swagger; DO NOT EDIT.

// Copyright 2022 Authors of spidernet-io
// SPDX-License-Identifier: Apache-2.0

package controller

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
)

// DeleteEndpointIDHandlerFunc turns a function with the right signature into a delete endpoint ID handler
type DeleteEndpointIDHandlerFunc func(DeleteEndpointIDParams) middleware.Responder

// Handle executing the request and returning a response
func (fn DeleteEndpointIDHandlerFunc) Handle(params DeleteEndpointIDParams) middleware.Responder {
	return fn(params)
}

// DeleteEndpointIDHandler interface for that can handle valid delete endpoint ID params
type DeleteEndpointIDHandler interface {
	Handle(DeleteEndpointIDParams) middleware.Responder
}

// NewDeleteEndpointID creates a new http.Handler for the delete endpoint ID operation
func NewDeleteEndpointID(ctx *middleware.Context, handler DeleteEndpointIDHandler) *DeleteEndpointID {
	return &DeleteEndpointID{Context: ctx, Handler: handler}
}

/*
	DeleteEndpointID swagger:route DELETE /endpoint/{id} controller deleteEndpointId

# Delete endpoint

Deletes the endpoint specified by the ID implemented by controller pod
*/
type DeleteEndpointID struct {
	Context *middleware.Context
	Handler DeleteEndpointIDHandler
}

func (o *DeleteEndpointID) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		*r = *rCtx
	}
	var Params = NewDeleteEndpointIDParams()
	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params) // actually handle the request
	o.Context.Respond(rw, r, route.Produces, route, res)

}