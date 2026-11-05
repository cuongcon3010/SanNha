package app

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestModuleBoundaries(t *testing.T) {
	root := filepath.Join("..", "modules")
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		module := strings.Split(filepath.ToSlash(rel), "/")[0]
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			return err
		}
		for _, imp := range file.Imports {
			value, _ := strconv.Unquote(imp.Path.Value)
			prefix := "sannha/be/internal/modules/"
			if strings.HasPrefix(value, prefix) {
				target := strings.TrimPrefix(value, prefix)
				if !(module == "booking" && target == "venue") {
					t.Errorf("forbidden module dependency: %s -> %s", module, target)
				}
			}
		}
		ast.Inspect(file, func(n ast.Node) bool {
			lit, ok := n.(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				return true
			}
			value, _ := strconv.Unquote(lit.Value)
			for owner, tables := range map[string][]string{"identity": {"identity_users", "identity_sessions", "identity_oauth_accounts", "identity_oauth_flows"}, "venue": {"venue_venues", "venue_courts", "venue_favorites"}, "booking": {"booking_bookings", "booking_payments", "booking_payment_events", "booking_callback_tokens", "booking_vouchers", "booking_voucher_uses"}} {
				for _, table := range tables {
					if owner != module && strings.Contains(value, table) {
						t.Errorf("%s references %s-owned storage in %s", module, owner, path)
					}
				}
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
