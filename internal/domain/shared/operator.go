package shared

import (
	"errors"
	"fmt"
)

// ErrOperatorRequired: an action that only staff may perform was attempted
// without saying who. Every context that asks "who did this" refuses with it
// (catalog's Provenance, procurement's receipt), so it lives here.
var ErrOperatorRequired = errors.New("this action needs the operator's id")

// OperatorID identifies a member of staff — the person who measured a parcel,
// confirmed a listing, bought at a shop, packed a box.
//
// "Who did this" is asked in every bounded context, so the IDENTITY lives in
// the shared kernel. Everything else about the person — name, role, login —
// does not: that is a future identity/access context, and other contexts
// refer to it by this id only.
//
// [PHP] Cùng mẫu với OrderID: nhúng ID để có String()/IsZero(), và là kiểu
// [PHP] riêng để không truyền nhầm OperatorID vào chỗ cần MerchantID.
// [PHP] Symfony: `final class OperatorId { public function __construct(public readonly Uuid $value) {} }`.
type OperatorID struct {
	ID
}

func NewOperatorID() OperatorID {
	return OperatorID{NewID()}
}

// ParseOperatorID reads an id back from a session, a URL or a database row.
func ParseOperatorID(s string) (OperatorID, error) {
	id, err := ParseID(s)
	if err != nil {
		return OperatorID{}, fmt.Errorf("operator id: %w", err)
	}
	return OperatorID{id}, nil
}
