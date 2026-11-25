package knowledge

import "github.com/zh7551069/qlutch-projects/services/api/internal/identity"

type accountView struct {
	UserID     string `json:"user_id"`
	IsNew      bool   `json:"is_new_user"`
	Onboarding bool   `json:"onboarding_completed"`
	FirstPack  bool   `json:"has_first_pack"`
}

type packView struct {
	PackID   string      `json:"pack_id"`
	Account  accountView `json:"account"`
	NextStep string      `json:"next_step"`
}

type uploadView struct {
	MaterialID string            `json:"material_id"`
	PackID     string            `json:"pack_id"`
	Source     string            `json:"source"`
	UploadURL  string            `json:"upload_url"`
	Method     string            `json:"upload_method"`
	Headers    map[string]string `json:"upload_headers"`
	ExpiresIn  int64             `json:"expires_in"`
	Account    accountView       `json:"account"`
	NextStep   string            `json:"next_step"`
}

type jobView struct {
	JobID       string      `json:"job_id"`
	PackID      string      `json:"pack_id"`
	MaterialID  string      `json:"material_id"`
	Source      string      `json:"source"`
	Status      string      `json:"status"`
	MapReady    bool        `json:"map_ready"`
	Retryable   bool        `json:"retryable"`
	Progress    *string     `json:"progress_message,omitempty"`
	FailureCode *string     `json:"failure_code"`
	FailureText *string     `json:"failure_message"`
	Account     accountView `json:"account"`
	NextStep    string      `json:"next_step"`
}

type skipView struct {
	Account  accountView `json:"account"`
	NextStep string      `json:"next_step"`
}

type packBody struct{}

type uploadBody struct {
	PackID      string  `json:"pack_id"`
	Source      string  `json:"source"`
	ContentType string  `json:"content_type"`
	ByteSize    int64   `json:"byte_size"`
	Filename    *string `json:"filename"`
}

type importBody struct {
	PackID     string `json:"pack_id"`
	MaterialID string `json:"material_id"`
}

type skipBody struct{}

type queuedJob struct {
	JobID string `json:"job_id"`
}

func accountOf(snap identity.AccountSnapshot) accountView {
	return accountView{
		UserID:     snap.UserID,
		IsNew:      snap.IsNewUser,
		Onboarding: snap.OnboardingCompleted,
		FirstPack:  snap.HasFirstPack,
	}
}

func textPtr(status, value string) *string {
	if status != statusFailedRetryable && status != statusFailedTerminal {
		return nil
	}
	if value == "" {
		return nil
	}
	copied := value
	return &copied
}

func progressPtr(value string) *string {
	if value == "" {
		return nil
	}
	copied := value
	return &copied
}
