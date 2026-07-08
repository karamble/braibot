// Copyright (c) 2026 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package mcpsrv

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/karamble/braibot/internal/database"
)

const (
	adminUID = "8cafda06372331b1fb712ab24382d890295cde558534072455046fb56399fb7a"
	dirUID   = "87df4e08e04b1c60e35fa41cf65e9e4a4c269e6c07709e10c136ba46e56683b5"
	plainUID = "76d2132c817d1bacad043855da6419c08076f1e153dbf362353d4aa3867298ed"
	otherUID = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
)

func newTestAdmin(t *testing.T, dataDir string, adminUIDs, dirUIDs []string) *Admin {
	t.Helper()
	a, err := NewAdmin(nil, dataDir, adminUIDs, dirUIDs)
	if err != nil {
		t.Fatalf("NewAdmin: %v", err)
	}
	return a
}

func TestNewAdminRejectsMalformedUID(t *testing.T) {
	if _, err := NewAdmin(nil, t.TempDir(), []string{"not-a-uid"}, nil); err == nil {
		t.Fatal("expected error for malformed admin uid")
	}
	if _, err := NewAdmin(nil, t.TempDir(), []string{adminUID[:63]}, nil); err == nil {
		t.Fatal("expected error for short admin uid")
	}
}

func TestToolVisible(t *testing.T) {
	a := newTestAdmin(t, t.TempDir(), []string{adminUID}, []string{dirUID})
	a.mu.Lock()
	a.adminTools["admin_ban"] = true
	a.mu.Unlock()

	cases := []struct {
		peer, tool string
		want       bool
	}{
		{adminUID, "admin_ban", true},
		{dirUID, "admin_ban", false},
		{plainUID, "admin_ban", false},
		{adminUID, "listing_invite", true},
		{dirUID, "listing_invite", true},
		{plainUID, "listing_invite", false},
		{adminUID, "text2image", true},
		{dirUID, "text2image", true},
		{plainUID, "text2image", true},
	}
	for _, c := range cases {
		if got := a.ToolVisible(c.peer, c.tool); got != c.want {
			t.Errorf("ToolVisible(%s, %s) = %v, want %v", c.peer[:8], c.tool, got, c.want)
		}
	}
}

func TestToolVisibleNoAdmins(t *testing.T) {
	a := newTestAdmin(t, t.TempDir(), nil, []string{dirUID})
	a.mu.Lock()
	a.adminTools["admin_ban"] = true
	a.mu.Unlock()

	if a.ToolVisible(plainUID, "admin_ban") {
		t.Error("admin tool visible with no admins configured")
	}
	if a.ToolVisible(plainUID, "listing_invite") {
		t.Error("listing_invite visible to a plain peer")
	}
	if !a.ToolVisible(dirUID, "listing_invite") {
		t.Error("listing_invite hidden from a configured directory")
	}
	if !a.ToolVisible(plainUID, "text2image") {
		t.Error("normal tool hidden from a plain peer")
	}
}

func TestAllowAndBanPersistence(t *testing.T) {
	dir := t.TempDir()
	a := newTestAdmin(t, dir, []string{adminUID}, nil)

	if !a.Allow(plainUID) {
		t.Fatal("unbanned peer refused")
	}
	a.mu.Lock()
	a.bans[plainUID] = banEntry{Note: "test", TS: 1}
	if err := a.saveBansLocked(); err != nil {
		a.mu.Unlock()
		t.Fatalf("saveBansLocked: %v", err)
	}
	a.mu.Unlock()

	if a.Allow(plainUID) {
		t.Error("banned peer admitted")
	}
	if !a.Allow(adminUID) {
		t.Error("admin refused")
	}

	// Re-open: the ban must survive.
	b := newTestAdmin(t, dir, []string{adminUID}, nil)
	if b.Allow(plainUID) {
		t.Error("ban lost across re-open")
	}

	// A stale ban on a configured admin is dropped at load.
	a.mu.Lock()
	a.bans[otherUID] = banEntry{TS: 2}
	if err := a.saveBansLocked(); err != nil {
		a.mu.Unlock()
		t.Fatalf("saveBansLocked: %v", err)
	}
	a.mu.Unlock()
	c := newTestAdmin(t, dir, []string{otherUID}, nil)
	if !c.Allow(otherUID) {
		t.Error("configured admin still banned after load")
	}
	if c.Allow(plainUID) {
		t.Error("unrelated ban dropped at load")
	}
}

func TestAdminLogShape(t *testing.T) {
	dir := t.TempDir()
	l := &adminLog{path: filepath.Join(dir, "adminlog.json")}
	l.append(adminUID, "admin_ban", map[string]string{"uid": plainUID})
	l.append(adminUID, "admin_unban", map[string]string{"uid": plainUID})

	raw, err := os.ReadFile(l.path)
	if err != nil {
		t.Fatalf("read adminlog: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(raw)), "\n")
	if len(lines) != 2 {
		t.Fatalf("got %d lines, want 2", len(lines))
	}
	var e adminLogEntry
	if err := json.Unmarshal([]byte(lines[0]), &e); err != nil {
		t.Fatalf("unmarshal entry: %v", err)
	}
	if e.Actor != adminUID || e.Action != "admin_ban" || e.TS == 0 {
		t.Errorf("unexpected entry: %+v", e)
	}
}

func TestValidUID(t *testing.T) {
	if !validUID(adminUID) {
		t.Error("valid uid rejected")
	}
	for _, bad := range []string{"", adminUID[:63], adminUID + "aa", strings.Repeat("z", 64)} {
		if validUID(bad) {
			t.Errorf("invalid uid %q accepted", bad)
		}
	}
}

func TestMatomsDCRConversion(t *testing.T) {
	matoms, err := dcrToMatoms(1.5)
	if err != nil {
		t.Fatalf("dcrToMatoms: %v", err)
	}
	if matoms != 150_000_000_000 {
		t.Errorf("dcrToMatoms(1.5) = %d, want 150000000000", matoms)
	}
	if dcr := matomsToDCR(matoms); dcr != 1.5 {
		t.Errorf("matomsToDCR round-trip = %v, want 1.5", dcr)
	}
}

func TestListBalances(t *testing.T) {
	db, err := database.NewDBManager(t.TempDir())
	if err != nil {
		t.Fatalf("NewDBManager: %v", err)
	}
	defer db.Close()

	if bals, err := db.ListBalances(); err != nil || len(bals) != 0 {
		t.Fatalf("empty db: got %v, %v", bals, err)
	}
	if err := db.UpdateBalance(plainUID, 1000); err != nil {
		t.Fatalf("UpdateBalance: %v", err)
	}
	if err := db.UpdateBalance(adminUID, 2000); err != nil {
		t.Fatalf("UpdateBalance: %v", err)
	}
	bals, err := db.ListBalances()
	if err != nil {
		t.Fatalf("ListBalances: %v", err)
	}
	if len(bals) != 2 {
		t.Fatalf("got %d balances, want 2", len(bals))
	}
	// ORDER BY uid: 76d2... sorts before 8caf...
	if bals[0].UID != plainUID || bals[0].Balance != 1000 {
		t.Errorf("unexpected first row: %+v", bals[0])
	}
	if bals[1].UID != adminUID || bals[1].Balance != 2000 {
		t.Errorf("unexpected second row: %+v", bals[1])
	}
}
