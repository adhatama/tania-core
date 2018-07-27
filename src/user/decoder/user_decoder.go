package decoder

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/Tanibox/tania-core/src/user/domain"
	"github.com/mitchellh/mapstructure"
)

type UserEventWrapper EventWrapper

func (w *UserEventWrapper) UnmarshalJSON(b []byte) error {
	wrapper := EventWrapper{}

	err := json.Unmarshal(b, &wrapper)
	if err != nil {
		return err
	}

	mapped, ok := wrapper.EventData.(map[string]interface{})
	if !ok {
		return errors.New("Error type assertion")
	}

	f := mapstructure.ComposeDecodeHookFunc(
		UIDHook(),
		TimeHook(time.RFC3339),
		PasswordHook(),
	)

	switch wrapper.EventName {
	case "UserAdminCreated":
		e := domain.UserAdminCreated{}

		_, err := Decode(f, &mapped, &e)
		if err != nil {
			return err
		}

		w.EventData = e

	case "UserInvited":
		e := domain.UserInvited{}

		_, err := Decode(f, &mapped, &e)
		if err != nil {
			return err
		}

		w.EventData = e

	case "PasswordChanged":
		e := domain.PasswordChanged{}

		_, err := Decode(f, &mapped, &e)
		if err != nil {
			return err
		}

		w.EventData = e

	case "UserVerified":
		e := domain.UserVerified{}

		_, err := Decode(f, &mapped, &e)
		if err != nil {
			return err
		}

		w.EventData = e

	case "ResetPasswordRequested":
		e := domain.ResetPasswordRequested{}

		_, err := Decode(f, &mapped, &e)
		if err != nil {
			return err
		}

		w.EventData = e

	case "InitialUserProfileSet":
		e := domain.InitialUserProfileSet{}

		_, err := Decode(f, &mapped, &e)
		if err != nil {
			return err
		}

		w.EventData = e

	}

	return nil
}
