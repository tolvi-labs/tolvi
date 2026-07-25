package vault

import (
	"path/filepath"
	"testing"
)

func TestResolveDocDestination(t *testing.T) {
	const pub = "/repo/vault"

	tests := []struct {
		name          string
		meta          Meta
		docType       string
		docVisibility string
		wantRoot      string
		wantRouted    bool
		wantErr       bool
	}{
		// --- Not public: everything stays local, for every doc type. ---
		{
			name:       "local session stays local",
			meta:       Meta{Workspace: "w"},
			docType:    "session",
			wantRoot:   pub,
			wantRouted: false,
		},
		{
			name:       "local decision stays local",
			meta:       Meta{Workspace: "w"},
			docType:    "decision",
			wantRoot:   pub,
			wantRouted: false,
		},
		{
			name:          "local private-marked decision still stays local",
			meta:          Meta{Workspace: "w"},
			docType:       "decision",
			docVisibility: "private",
			wantRoot:      pub,
			wantRouted:    false,
		},
		{
			name:       "unknown visibility value behaves as local",
			meta:       Meta{Workspace: "w", Visibility: "internal"},
			docType:    "session",
			wantRoot:   pub,
			wantRouted: false,
		},

		// --- Public: sessions always route. ---
		{
			name:       "public session routes to absolute private vault",
			meta:       Meta{Workspace: "w", Visibility: "public", PrivateVault: "/other/vault"},
			docType:    "session",
			wantRoot:   "/other/vault",
			wantRouted: true,
		},
		{
			name:       "public session routes to relative private vault (resolved against public)",
			meta:       Meta{Workspace: "w", Visibility: "public", PrivateVault: "../private/vault"},
			docType:    "session",
			wantRoot:   filepath.Join(pub, "../private/vault"),
			wantRouted: true,
		},
		{
			name:    "public session with no private_vault errors",
			meta:    Meta{Workspace: "w", Visibility: "public"},
			docType: "session",
			wantErr: true,
		},

		// --- Public: decisions/patterns stay public unless marked private. ---
		{
			name:       "public decision (default visibility) stays public",
			meta:       Meta{Workspace: "w", Visibility: "public", PrivateVault: "/other/vault"},
			docType:    "decision",
			wantRoot:   pub,
			wantRouted: false,
		},
		{
			name:       "public pattern (default visibility) stays public",
			meta:       Meta{Workspace: "w", Visibility: "public", PrivateVault: "/other/vault"},
			docType:    "pattern",
			wantRoot:   pub,
			wantRouted: false,
		},
		{
			name:          "public private-marked decision routes to private",
			meta:          Meta{Workspace: "w", Visibility: "public", PrivateVault: "/other/vault"},
			docType:       "decision",
			docVisibility: "private",
			wantRoot:      "/other/vault",
			wantRouted:    true,
		},
		{
			name:          "public private-marked decision with no private_vault errors",
			meta:          Meta{Workspace: "w", Visibility: "public"},
			docType:       "decision",
			docVisibility: "private",
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, routed, err := ResolveDocDestination(pub, tt.meta, tt.docType, tt.docVisibility)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got root=%q routed=%v", root, routed)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if root != tt.wantRoot {
				t.Errorf("root = %q, want %q", root, tt.wantRoot)
			}
			if routed != tt.wantRouted {
				t.Errorf("routed = %v, want %v", routed, tt.wantRouted)
			}
		})
	}
}
