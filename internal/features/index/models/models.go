package models

type ReactionRequest struct {
	Reaction string `json:"reaction"`
}

func (r ReactionRequest) IsValid() bool {
	return r.Reaction == "like" || r.Reaction == "love" || r.Reaction == "laugh"
}
