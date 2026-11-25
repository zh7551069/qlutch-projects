package identity

type sendView struct {
	RetryAfter int64  `json:"retry_after_seconds"`
	Hint       string `json:"destination_hint"`
}

type sessionView struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	ExpiresIn    int64       `json:"expires_in"`
	Account      accountView `json:"account"`
	NextStep     string      `json:"next_step"`
	Profile      profileView `json:"profile"`
}

type accountView struct {
	UserID     string `json:"user_id"`
	IsNew      bool   `json:"is_new_user"`
	Onboarding bool   `json:"onboarding_completed"`
	FirstPack  bool   `json:"has_first_pack"`
}

type profileView struct {
	Locale string `json:"locale"`
	Region string `json:"region_code"`
}

type onboardingView struct {
	CurrentStepID *string     `json:"current_step_id"`
	CurrentIndex  int64       `json:"current_index"`
	Completed     bool        `json:"completed"`
	Nickname      *string     `json:"nickname"`
	Steps         []stepView  `json:"steps"`
	Account       accountView `json:"account"`
	NextStep      string      `json:"next_step"`
}

type stepView struct {
	StepID string  `json:"step_id"`
	Index  int64   `json:"index"`
	Status string  `json:"status"`
	Answer *string `json:"answer"`
}
