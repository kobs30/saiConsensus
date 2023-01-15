package models

import valid "github.com/asaskevich/govalidator"

// AddValidator request struct
type AddValidatorRequest struct {
	AddressToAdd string `json:"address_to_add" valid:",required"`
	AddressFrom  string `json:"address_from" valid:",required"`
	Signature    string `json:"signature" valid:",required"`
}

// Validate transaction message
func (m *AddValidatorRequest) Validate() error {
	_, err := valid.ValidateStruct(m)
	return err
}
