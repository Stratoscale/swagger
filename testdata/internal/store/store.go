package store

import (
	"context"

	"github.com/Stratoscale/swagger/testdata/restapi/operations/store"
	"github.com/go-openapi/runtime/middleware"
)

type Store struct{}

func (*Store) InventoryGet(ctx context.Context, params store.InventoryGetParams) middleware.Responder {
	panic("implement me")
}

func (*Store) OrderCreate(ctx context.Context, params store.OrderCreateParams) middleware.Responder {
	panic("implement me")
}

func (*Store) OrderDelete(ctx context.Context, params store.OrderDeleteParams) middleware.Responder {
	panic("implement me")
}

func (*Store) OrderGet(ctx context.Context, params store.OrderGetParams) middleware.Responder {
	panic("implement me")
}
