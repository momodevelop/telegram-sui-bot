package lta

type API struct {
	Token string
}

func New(token string) *API {
	ret := &API{
		Token: token,
	}
	return ret
}
