package logincheck



/*
func TestSuccessfulLogin(t *testing.T) {
	u := &auth.User{Username:"admin", Password: utils.HashPassword("123456")}
	config.DB.Create(u)


	w := utils.HostEndpoint(ProcessRequest, `{ "username": "admin", "password": "123456" }`)

	if w.Code != http.StatusOK {
		t.Errorf("Status code should be %v, was %d", http.StatusOK, w.Code)
	}

	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Errorf("Expected one cookie, got %d", len(cookies))
	}

	a := storage.AuthDb{}
	sessionId :=  cookies[0].Value
	session := a.Load(sessionId)
	if session == nil {
		t.Error("Expected session to exist in the database")
	}
}

func TestInexistingUserShouldReturnUnauthorized(t *testing.T) {
	w := utils.HostEndpoint(ProcessRequest, `{ "username": "dontexist", "password": "dontexist" }`)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Status code should be %v, was %d", http.StatusUnauthorized, w.Code)
	}
}

func TestWrongPasswordShouldReturnUnauthorized(t *testing.T) {
	w := utils.HostEndpoint(ProcessRequest, `{ "username": "admin", "password": "wrongpass" }`)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Status code should be %v, was %d", http.StatusUnauthorized, w.Code)
	}
}
*/