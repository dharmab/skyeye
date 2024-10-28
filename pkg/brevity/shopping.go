package brevity

// Shopping is an air-to-ground brevity commonly mistaken as an air-to-air brevity.
type ShoppingRequest struct {
	Callsign string
}

func (r ShoppingRequest) String() string {
	return "SHOPPING for " + r.Callsign
}

// ShoppingResponse is reeducation.
type ShoppingResponse struct {
	Callsign string
}
