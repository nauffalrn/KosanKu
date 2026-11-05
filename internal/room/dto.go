package room

import "time"

// DTO (Data Transfer Object) memisahkan bentuk JSON dari entity.
// Entity bisa berubah tanpa merusak kontrak API, dan sebaliknya.

type createRequest struct {
	Number       string `json:"number"`
	MonthlyPrice int64  `json:"monthly_price"`
}

func (req createRequest) toInput() CreateInput {
	// Konversi langsung bisa dilakukan karena field-nya identik (tag JSON diabaikan).
	// Jika nanti field-nya berbeda, compiler akan memberi tahu.
	return CreateInput(req)
}

type updateRequest struct {
	Number       *string `json:"number"`
	MonthlyPrice *int64  `json:"monthly_price"`
	Status       *string `json:"status"`
}

func (req updateRequest) toInput() UpdateInput {
	in := UpdateInput{Number: req.Number, MonthlyPrice: req.MonthlyPrice}
	if req.Status != nil {
		status := Status(*req.Status)
		in.Status = &status
	}
	return in
}

type response struct {
	ID           string    `json:"id"`
	Number       string    `json:"number"`
	MonthlyPrice int64     `json:"monthly_price"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func toResponse(r Room) response {
	return response{
		ID:           r.ID,
		Number:       r.Number,
		MonthlyPrice: r.MonthlyPrice,
		Status:       string(r.Status),
		CreatedAt:    r.CreatedAt,
		UpdatedAt:    r.UpdatedAt,
	}
}

func toResponses(rooms []Room) []response {
	out := make([]response, 0, len(rooms))
	for _, r := range rooms {
		out = append(out, toResponse(r))
	}
	return out
}
