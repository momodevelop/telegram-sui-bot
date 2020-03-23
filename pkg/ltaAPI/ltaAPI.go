


type LtaAPI struct{
	Token string
}

func NewLtaAPI(token string) *LtaAPI {
	ret := &LtaAPI{
		Token: token
	}
	return ret;
}

func Call(path string) interface{} {
	req := "http://datamall2.mytransport.sg" + path
}


