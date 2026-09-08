package logistics

import "errors"

var (
	ErrInvalidParcel         = errors.New("invalid parcel")
	ErrParcelNotExpected     = errors.New("parcel is not awaiting receipt")
	ErrParcelNotReceived     = errors.New("parcel has not been received")
	ErrParcelNotBatched      = errors.New("parcel is not in a batch")
	ErrParcelNotFound        = errors.New("parcel not found")
	ErrParcelAlreadyExpected = errors.New("a parcel is already expected for this order")
	ErrInvalidBatch          = errors.New("invalid batch")
	ErrBatchNotOpen          = errors.New("batch is not open")
	ErrBatchNotClosed        = errors.New("batch is not closed")
	ErrBatchEmpty            = errors.New("batch has no parcels")
	ErrDuplicateParcel       = errors.New("parcel already in the batch")
	ErrBatchNotFound         = errors.New("batch not found")
	ErrLaneRuleNotFound      = errors.New("lane rule not found")
	ErrInvalidFreight        = errors.New("freight must be a positive amount")
	ErrInvalidSnapshot       = errors.New("invalid logistics snapshot")
)
