package web

import (
	"os"
	"strings"
	"testing"
)

func TestAdminRedirectsExpiredTokensWithVisibleReason(t *testing.T) {
	client, err := os.ReadFile("admin/src/api/client.ts")
	if err != nil {
		t.Fatalf("read admin API client: %v", err)
	}
	login, err := os.ReadFile("admin/src/pages/Login.tsx")
	if err != nil {
		t.Fatalf("read login page: %v", err)
	}

	for _, want := range []string{
		`res.status === 401`,
		`tokenStore.clear()`,
		`reason=session-expired`,
	} {
		if !strings.Contains(string(client), want) {
			t.Errorf("admin API client missing expired-token behavior %q", want)
		}
	}
	for _, want := range []string{
		`session-expired`,
		`登录状态已失效，请重新登录`,
	} {
		if !strings.Contains(string(login), want) {
			t.Errorf("admin login page missing expired-token notice %q", want)
		}
	}
}

func TestDashboardDoesNotPresentFailedPostRequestAsZero(t *testing.T) {
	source, err := os.ReadFile("admin/src/pages/Dashboard.tsx")
	if err != nil {
		t.Fatalf("read admin dashboard: %v", err)
	}

	for _, want := range []string{
		`isError`,
		`文章数据加载失败`,
		`isError ? '—'`,
	} {
		if !strings.Contains(string(source), want) {
			t.Errorf("admin dashboard missing request-error behavior %q", want)
		}
	}
}
