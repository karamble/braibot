// Copyright (c) 2026 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package mcpsrv

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/decred/dcrd/dcrutil/v4"
	"github.com/karamble/brmcp/directory"
	"github.com/karamble/brmcp/server"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/karamble/braibot/internal/database"
)

// Admin gates the harness for operator control: admin-only tools hidden
// from everyone else, a persisted ban list enforced at envelope admission,
// and the listing_invite visibility restriction. The admin and directory
// uid sets are fixed at construction because the harness consults
// ToolVisible only once per peer.
type Admin struct {
	admins  map[string]bool
	dirUIDs map[string]bool
	db      *database.DBManager
	dataDir string
	alog    *adminLog

	mu         sync.Mutex
	adminTools map[string]bool
	bans       map[string]banEntry
	reg        *directory.Registrant
	autofund   directory.AutoFund
	haveDir    bool
}

type banEntry struct {
	Note string `json:"note,omitempty"`
	TS   int64  `json:"ts"`
}

// NewAdmin builds the admin gate from the adminuids and directoryuids
// config values and loads the persisted ban list. A malformed admin uid is
// a startup error; configured admins are unbannable, so any stale ban on
// one is dropped at load.
func NewAdmin(db *database.DBManager, dataDir string, adminUIDs, dirUIDs []string) (*Admin, error) {
	a := &Admin{
		admins:     make(map[string]bool),
		dirUIDs:    make(map[string]bool),
		db:         db,
		dataDir:    dataDir,
		adminTools: make(map[string]bool),
		bans:       make(map[string]banEntry),
	}
	for _, uid := range adminUIDs {
		u := strings.ToLower(strings.TrimSpace(uid))
		if !validUID(u) {
			return nil, fmt.Errorf("adminuids entry %q is not a 64-hex Bison Relay uid", uid)
		}
		a.admins[u] = true
	}
	for _, uid := range dirUIDs {
		a.dirUIDs[strings.ToLower(strings.TrimSpace(uid))] = true
	}
	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		return nil, err
	}
	raw, err := os.ReadFile(a.bansPath())
	switch {
	case errors.Is(err, os.ErrNotExist):
	case err != nil:
		return nil, err
	default:
		if err := json.Unmarshal(raw, &a.bans); err != nil {
			return nil, fmt.Errorf("ban list %s corrupt: %w", a.bansPath(), err)
		}
		for uid := range a.bans {
			if a.admins[uid] {
				delete(a.bans, uid)
			}
		}
	}
	a.alog = &adminLog{path: filepath.Join(dataDir, "adminlog.json")}
	return a, nil
}

func validUID(uid string) bool {
	if len(uid) != 64 {
		return false
	}
	_, err := hex.DecodeString(uid)
	return err == nil
}

func (a *Admin) bansPath() string { return filepath.Join(a.dataDir, "bans.json") }

func (a *Admin) isAdmin(uid string) bool { return a.admins[strings.ToLower(uid)] }

// Allow is the harness AllowFunc: admins always pass, everyone else passes
// unless banned. Consulted on every inbound envelope, so bans apply
// immediately without a restart.
func (a *Admin) Allow(uid string) bool {
	u := strings.ToLower(uid)
	if a.admins[u] {
		return true
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	_, banned := a.bans[u]
	return !banned
}

// ToolVisible is the harness visibility hook: admin tools exist only for
// admins, and listing_invite only for admins and the configured
// directories (its legitimate caller).
func (a *Admin) ToolVisible(peer, tool string) bool {
	p := strings.ToLower(peer)
	if tool == "listing_invite" {
		return a.admins[p] || a.dirUIDs[p]
	}
	a.mu.Lock()
	isAdminTool := a.adminTools[tool]
	a.mu.Unlock()
	return !isAdminTool || a.admins[p]
}

// SetDirectory hands the admin gate the registrant and the policy it was
// started with; the registrant keeps its config private, so admin_autofund
// reports this snapshot.
func (a *Admin) SetDirectory(reg *directory.Registrant, policy directory.AutoFund) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.reg = reg
	a.autofund = policy
	a.haveDir = true
}

func (a *Admin) saveBansLocked() error {
	raw, err := json.MarshalIndent(a.bans, "", "  ")
	if err != nil {
		return err
	}
	tmp := a.bansPath() + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, a.bansPath())
}

// adminLog is the append-only JSONL audit trail of every mutating admin
// call; an audit log must never be rewritten.
type adminLog struct {
	mu   sync.Mutex
	path string
}

type adminLogEntry struct {
	TS     int64  `json:"ts"`
	Actor  string `json:"actor"`
	Action string `json:"action"`
	Args   any    `json:"args,omitempty"`
}

func (l *adminLog) append(actor, action string, args any) {
	raw, err := json.Marshal(adminLogEntry{
		TS: time.Now().Unix(), Actor: actor, Action: action, Args: args,
	})
	if err != nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	f, err := os.OpenFile(l.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = f.Write(append(raw, '\n'))
}

// adminTool registers a tool that only admin uids can see or call. The
// visibility gate hides it from everyone else; the handler check is the
// belt to that suspender.
func adminTool[In any](a *Admin, h *server.Harness, name, desc string,
	fn func(ctx context.Context, actor string, in In) (any, error)) {

	a.mu.Lock()
	a.adminTools[name] = true
	a.mu.Unlock()
	server.AddTool(h, &mcp.Tool{Name: name, Description: desc}, 0,
		func(ctx context.Context, peer string, in In) (any, error) {
			if !a.isAdmin(peer) {
				return nil, errors.New("admin only")
			}
			return fn(ctx, peer, in)
		})
}

func matomsToDCR(matoms int64) float64 {
	return float64(matoms) / (matomsPerAtom * dcrutil.AtomsPerCoin)
}

func dcrToMatoms(dcr float64) (int64, error) {
	amt, err := dcrutil.NewAmount(dcr)
	if err != nil {
		return 0, err
	}
	return int64(amt) * matomsPerAtom, nil
}

type adminUIDIn struct {
	UID string `json:"uid" jsonschema:"the user's 64-hex Bison Relay uid"`
}

type adminAmountIn struct {
	UID       string  `json:"uid" jsonschema:"the user's 64-hex Bison Relay uid"`
	AmountDcr float64 `json:"amount_dcr" jsonschema:"amount in DCR"`
}

type adminBanIn struct {
	UID  string `json:"uid" jsonschema:"the peer's 64-hex Bison Relay uid"`
	Note string `json:"note,omitempty" jsonschema:"optional reason recorded with the ban"`
}

type adminRegisterIn struct {
	DirectoryUID string `json:"directory_uid" jsonschema:"the directory's 64-hex Bison Relay uid"`
}

// fundEntry mirrors the registrant's regfund.json records so
// admin_autofund can report the rolling 30-day spend.
type fundEntry struct {
	TS           int64  `json:"ts"`
	DirectoryUID string `json:"directoryUid"`
	Atoms        int64  `json:"atoms"`
}

// AttachAdmin registers the operator tools on the harness. Call it before
// any peer connects so the visibility predicate is stable.
func (a *Admin) AttachAdmin(h *server.Harness) {
	adminTool(a, h, "admin_balances",
		"List every user balance in DCR.",
		func(_ context.Context, _ string, _ struct{}) (any, error) {
			balances, err := a.db.ListBalances()
			if err != nil {
				return nil, err
			}
			users := make([]map[string]any, 0, len(balances))
			var total float64
			for _, b := range balances {
				dcr := matomsToDCR(b.Balance)
				total += dcr
				users = append(users, map[string]any{"uid": b.UID, "balance_dcr": dcr})
			}
			return map[string]any{"users": users, "count": len(users), "total_dcr": total}, nil
		})

	adminTool(a, h, "admin_credit",
		"Credit a user's balance by an amount in DCR. An unknown uid gets a fresh balance row.",
		func(_ context.Context, actor string, in adminAmountIn) (any, error) {
			uid := strings.ToLower(in.UID)
			if !validUID(uid) {
				return nil, errors.New("uid must be 64 hex characters")
			}
			if in.AmountDcr <= 0 {
				return nil, errors.New("amount_dcr must be positive")
			}
			matoms, err := dcrToMatoms(in.AmountDcr)
			if err != nil {
				return nil, err
			}
			if err := a.db.UpdateBalance(uid, matoms); err != nil {
				return nil, err
			}
			a.alog.append(actor, "admin_credit", in)
			bal, err := a.db.GetBalance(uid)
			if err != nil {
				return nil, err
			}
			return map[string]any{"uid": uid, "balance_dcr": matomsToDCR(bal)}, nil
		})

	adminTool(a, h, "admin_debit",
		"Debit a user's balance by an amount in DCR. Refused when it would drive the balance below zero.",
		func(_ context.Context, actor string, in adminAmountIn) (any, error) {
			uid := strings.ToLower(in.UID)
			if !validUID(uid) {
				return nil, errors.New("uid must be 64 hex characters")
			}
			if in.AmountDcr <= 0 {
				return nil, errors.New("amount_dcr must be positive")
			}
			matoms, err := dcrToMatoms(in.AmountDcr)
			if err != nil {
				return nil, err
			}
			raw, err := hex.DecodeString(uid)
			if err != nil {
				return nil, err
			}
			if ok, err := a.db.CheckAndDeductBalance(raw, matoms, false); !ok {
				return nil, err
			}
			a.alog.append(actor, "admin_debit", in)
			bal, err := a.db.GetBalance(uid)
			if err != nil {
				return nil, err
			}
			return map[string]any{"uid": uid, "balance_dcr": matomsToDCR(bal)}, nil
		})

	adminTool(a, h, "admin_ban",
		"Ban a peer uid from the MCP service. Takes effect immediately; admins cannot be banned.",
		func(_ context.Context, actor string, in adminBanIn) (any, error) {
			uid := strings.ToLower(in.UID)
			if !validUID(uid) {
				return nil, errors.New("uid must be 64 hex characters")
			}
			if a.admins[uid] {
				return nil, errors.New("cannot ban an admin")
			}
			a.mu.Lock()
			old, existed := a.bans[uid]
			a.bans[uid] = banEntry{Note: in.Note, TS: time.Now().Unix()}
			if err := a.saveBansLocked(); err != nil {
				if existed {
					a.bans[uid] = old
				} else {
					delete(a.bans, uid)
				}
				a.mu.Unlock()
				return nil, err
			}
			a.mu.Unlock()
			a.alog.append(actor, "admin_ban", in)
			return map[string]any{"uid": uid, "banned": true}, nil
		})

	adminTool(a, h, "admin_unban",
		"Lift a peer uid's ban.",
		func(_ context.Context, actor string, in adminUIDIn) (any, error) {
			uid := strings.ToLower(in.UID)
			if !validUID(uid) {
				return nil, errors.New("uid must be 64 hex characters")
			}
			a.mu.Lock()
			old, existed := a.bans[uid]
			if existed {
				delete(a.bans, uid)
				if err := a.saveBansLocked(); err != nil {
					a.bans[uid] = old
					a.mu.Unlock()
					return nil, err
				}
			}
			a.mu.Unlock()
			if existed {
				a.alog.append(actor, "admin_unban", in)
			}
			return map[string]any{"uid": uid, "banned": false, "was_banned": existed}, nil
		})

	adminTool(a, h, "admin_list_bans",
		"List the banned peer uids.",
		func(_ context.Context, _ string, _ struct{}) (any, error) {
			a.mu.Lock()
			bans := make([]map[string]any, 0, len(a.bans))
			for uid, e := range a.bans {
				bans = append(bans, map[string]any{"uid": uid, "note": e.Note, "ts": e.TS})
			}
			a.mu.Unlock()
			sort.Slice(bans, func(i, j int) bool {
				return bans[i]["uid"].(string) < bans[j]["uid"].(string)
			})
			return map[string]any{"bans": bans}, nil
		})

	adminTool(a, h, "admin_register",
		"Register (or renew) this bot's listing at a directory. Funds the fee under the AutoFund policy.",
		func(ctx context.Context, actor string, in adminRegisterIn) (any, error) {
			uid := strings.ToLower(in.DirectoryUID)
			if !validUID(uid) {
				return nil, errors.New("directory_uid must be 64 hex characters")
			}
			a.mu.Lock()
			reg := a.reg
			a.mu.Unlock()
			if reg == nil {
				return nil, errors.New("directory support is not enabled")
			}
			if err := reg.Register(ctx, uid); err != nil {
				return nil, err
			}
			a.alog.append(actor, "admin_register", in)
			return map[string]any{"directory_uid": uid, "registered": true}, nil
		})

	adminTool(a, h, "admin_autofund",
		"Show the directory auto-fund policy and the rolling 30-day listing spend.",
		func(_ context.Context, _ string, _ struct{}) (any, error) {
			a.mu.Lock()
			haveDir, policy := a.haveDir, a.autofund
			a.mu.Unlock()
			if !haveDir {
				return nil, errors.New("directory support is not enabled")
			}
			var entries []fundEntry
			raw, err := os.ReadFile(filepath.Join(a.dataDir, "regfund.json"))
			switch {
			case errors.Is(err, os.ErrNotExist):
			case err != nil:
				return nil, err
			default:
				if err := json.Unmarshal(raw, &entries); err != nil {
					return nil, err
				}
			}
			cutoff := time.Now().Add(-30 * 24 * time.Hour).Unix()
			var spent int64
			for _, e := range entries {
				if e.TS >= cutoff {
					spent += e.Atoms
				}
			}
			return map[string]any{
				"enabled":                policy.Enabled,
				"max_atoms_per_request":  policy.MaxAtomsPerRequest,
				"max_atoms_per_month":    policy.MaxAtomsPerMonth,
				"allowed_directory_uids": policy.AllowedDirectoryUIDs,
				"spent_30d_atoms":        spent,
				"note":                   "policy is set in braibot.conf and applies after restart",
			}, nil
		})
}
