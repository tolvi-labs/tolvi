package vault

import (
	"fmt"
	"path/filepath"
)

// ResolveDocDestination decides which vault root a new doc should be written
// to when the public/private routing feature is active.
//
// publicVaultPath is the repo's own (public) vault. m is that vault's meta.
// docType is "decision" | "session" | "pattern". docVisibility is the doc's
// own visibility ("private" marks a decision/pattern as internal; empty is
// the default).
//
// Rules:
//   - m.Visibility != "public": returns (publicVaultPath, false, nil) — the
//     doc stays local, for every doc type (unchanged legacy behavior).
//   - m.Visibility == "public":
//   - sessions ALWAYS route to the private vault.
//   - decisions/patterns route to the private vault ONLY when
//     docVisibility=="private"; otherwise they stay in publicVaultPath.
//   - When routing to the private vault, m.PrivateVault must be set (else an
//     error is returned). A relative m.PrivateVault is resolved against
//     publicVaultPath; an absolute one is used as-is.
func ResolveDocDestination(publicVaultPath string, m Meta, docType, docVisibility string) (root string, routedToPrivate bool, err error) {
	if m.Visibility != "public" {
		return publicVaultPath, false, nil
	}
	routes := docType == "session" || docVisibility == "private"
	if !routes {
		return publicVaultPath, false, nil
	}
	if m.PrivateVault == "" {
		return "", false, fmt.Errorf(`visibility is "public" but private_vault is not configured`)
	}
	priv := m.PrivateVault
	if !filepath.IsAbs(priv) {
		priv = filepath.Join(publicVaultPath, priv)
	}
	return priv, true, nil
}
