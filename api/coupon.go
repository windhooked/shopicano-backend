package api

import (
	"github.com/labstack/echo/v4"
	"github.com/shopicano/shopicano-backend/app"
	"github.com/shopicano/shopicano-backend/core"
	"github.com/shopicano/shopicano-backend/data"
	"github.com/shopicano/shopicano-backend/errors"
	"github.com/shopicano/shopicano-backend/middlewares"
	"github.com/shopicano/shopicano-backend/models"
	"github.com/shopicano/shopicano-backend/utils"
	"github.com/shopicano/shopicano-backend/validators"
	"net/http"
	"strconv"
	"time"
)

func RegisterCouponRoutes(g *echo.Group) {
	func(g *echo.Group) {
		// Private endpoints only
		g.Use(middlewares.IsStoreStaffWithStoreActivation)
		g.POST("/", createCoupon)
		g.PATCH("/:coupon_id/", updateCoupon)
		g.DELETE("/:coupon_id/", deleteCoupon)
		g.GET("/:coupon_id/", getCoupon)
		g.GET("/", listCoupons)
	}(g)

	func(g *echo.Group) {
		// Private endpoints only
		g.Use(middlewares.AuthUser)
		g.POST("/:coupon_id/", checkCouponAvailability)
	}(g)
}

func createCoupon(ctx echo.Context) error {
	storeID := ctx.Get(utils.StoreID).(string)

	req, err := validators.ValidateCreateCoupon(ctx)

	resp := core.Response{}

	if err != nil {
		resp.Title = "Invalid data"
		resp.Status = http.StatusUnprocessableEntity
		resp.Code = errors.ProductCreationDataInvalid
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	db := app.DB()
	cr := data.NewCouponRepository()

	st, _ := utils.ParseDateTimeForInput(req.StartAt)
	et, _ := utils.ParseDateTimeForInput(req.EndAt)

	c := models.Coupon{
		ID:             utils.NewUUID(),
		StoreID:        storeID,
		IsUserSpecific: req.IsUserSpecific,
		DiscountType:   req.DiscountType,
		Code:           req.Code,
		DiscountAmount: req.DiscountAmount,
		StartAt:        st,
		EndAt:          et,
		IsActive:       req.IsActive,
		IsFlatDiscount: req.IsFlatDiscount,
		MaxDiscount:    req.MaxDiscount,
		MaxUsage:       req.MaxUsage,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}

	err = cr.Create(db, &c)
	if err != nil {
		msg, ok := errors.IsDuplicateKeyError(err)
		if ok {
			resp.Title = msg
			resp.Status = http.StatusConflict
			resp.Code = errors.CouponAlreadyExists
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}

		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusCreated
	resp.Title = "Coupon created"
	resp.Data = c
	return resp.ServerJSON(ctx)
}

func updateCoupon(ctx echo.Context) error {
	storeID := ctx.Get(utils.StoreID).(string)
	couponID := ctx.Param("coupon_id")

	req, err := validators.ValidateUpdateCoupon(ctx)

	resp := core.Response{}

	if err != nil {
		resp.Title = "Invalid data"
		resp.Status = http.StatusUnprocessableEntity
		resp.Code = errors.ProductCreationDataInvalid
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	db := app.DB()
	cr := data.NewCouponRepository()

	c, err := cr.Get(db, storeID, couponID)
	if err != nil {
		if errors.IsRecordNotFoundError(err) {
			resp.Title = "Coupon not found"
			resp.Status = http.StatusNotFound
			resp.Code = errors.CouponNotFound
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}

		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	if req.Code != nil {
		c.Code = *req.Code
	}
	if req.IsUserSpecific != nil {
		c.IsUserSpecific = *req.IsUserSpecific
	}
	if req.DiscountType != nil {
		c.DiscountType = *req.DiscountType
	}
	if req.DiscountAmount != nil {
		c.DiscountAmount = *req.DiscountAmount
	}
	if req.IsActive != nil {
		c.IsActive = *req.IsActive
	}
	if req.IsFlatDiscount != nil {
		c.IsFlatDiscount = *req.IsFlatDiscount
	}
	if req.MaxDiscount != nil {
		c.MaxDiscount = *req.MaxDiscount
	}
	if req.MaxUsage != nil {
		c.MaxUsage = *req.MaxUsage
	}
	if req.StartAt != nil {
		st, _ := utils.ParseDateTimeForInput(*req.StartAt)
		c.StartAt = st
	}
	if req.EndAt != nil {
		et, _ := utils.ParseDateTimeForInput(*req.EndAt)
		c.EndAt = et
	}

	c.UpdatedAt = time.Now().UTC()

	err = cr.Update(db, c)
	if err != nil {
		msg, ok := errors.IsDuplicateKeyError(err)
		if ok {
			resp.Title = msg
			resp.Status = http.StatusConflict
			resp.Code = errors.CouponAlreadyExists
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}

		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusOK
	resp.Title = "Coupon updated"
	resp.Data = c
	return resp.ServerJSON(ctx)
}

func deleteCoupon(ctx echo.Context) error {
	storeID := ctx.Get(utils.StoreID).(string)
	couponID := ctx.Param("coupon_id")

	resp := core.Response{}

	db := app.DB()

	cr := data.NewCouponRepository()
	if err := cr.Delete(db, storeID, couponID); err != nil {
		if errors.IsRecordNotFoundError(err) {
			resp.Title = "Coupon not found"
			resp.Status = http.StatusNotFound
			resp.Code = errors.CouponNotFound
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}

		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusNoContent
	return resp.ServerJSON(ctx)
}

func getCoupon(ctx echo.Context) error {
	storeID := ctx.Get(utils.StoreID).(string)
	couponID := ctx.Param("coupon_id")

	resp := core.Response{}

	db := app.DB()

	cr := data.NewCouponRepository()
	v, err := cr.Get(db, storeID, couponID)
	if err != nil {
		if errors.IsRecordNotFoundError(err) {
			resp.Title = "Coupon not found"
			resp.Status = http.StatusNotFound
			resp.Code = errors.CouponNotFound
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}

		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusOK
	resp.Data = v
	return resp.ServerJSON(ctx)
}

func listCoupons(ctx echo.Context) error {
	pageQ := ctx.Request().URL.Query().Get("page")
	limitQ := ctx.Request().URL.Query().Get("limit")
	query := ctx.Request().URL.Query().Get("query")

	page, err := strconv.ParseInt(pageQ, 10, 64)
	if err != nil {
		page = 1
	}
	limit, err := strconv.ParseInt(limitQ, 10, 64)
	if err != nil {
		limit = 10
	}

	resp := core.Response{}

	var r interface{}

	if query == "" {
		r, err = fetchCoupons(ctx, page, limit)
	} else {
		r, err = searchCoupons(ctx, query, page, limit)
	}

	if err != nil {
		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusOK
	resp.Data = r
	return resp.ServerJSON(ctx)
}

func fetchCoupons(ctx echo.Context, page int64, limit int64) (interface{}, error) {
	from := (page - 1) * limit
	cr := data.NewCouponRepository()
	db := app.DB()
	return cr.List(db, ctx.Get(utils.StoreID).(string), int(from), int(limit))
}

func searchCoupons(ctx echo.Context, query string, page int64, limit int64) (interface{}, error) {
	from := (page - 1) * limit
	cr := data.NewCouponRepository()
	db := app.DB()
	return cr.Search(db, ctx.Get(utils.StoreID).(string), query, int(from), int(limit))
}

func checkCouponAvailability(ctx echo.Context) error {
	productID := ctx.Param("product_id")

	resp := core.Response{}

	db := app.DB()

	pu := data.NewProductRepository()

	var p interface{}
	var err error

	if utils.IsStoreStaff(ctx) {
		p, err = pu.GetAsStoreStuff(db, ctx.Get(utils.StoreID).(string), productID)
	} else {
		p, err = pu.GetDetails(db, productID)
	}

	if err != nil {
		if errors.IsRecordNotFoundError(err) {
			resp.Title = "Product not found"
			resp.Status = http.StatusNotFound
			resp.Code = errors.ProductNotFound
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}

		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusOK
	resp.Data = p
	return resp.ServerJSON(ctx)
}