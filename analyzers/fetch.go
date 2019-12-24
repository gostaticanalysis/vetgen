package analyzers

import (
	"context"
	"errors"
	"fmt"
	"go/types"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/sync/errgroup"
	"golang.org/x/tools/go/packages"
)

const defaultBaseURL = "https://pkg.go.dev/"

var ignores = []string{
	"git.k2software.com.cn/go/",
	"gitea.com/xgo/",
	"gitee.com/wangHvip/",
	"github.com/aaronbee/tools/",
	"github.com/codeactual/",
	"github.com/determined-ai/tools/",
	"github.com/gcc-mirror/gcc/",
	"github.com/golang/",
	"github.com/golangci/",
	"github.com/heschik/tools/",
	"github.com/kamilsk/go-tools/",
	"github.com/lhecker/tools/",
	"github.com/mewpull/tools/",
	"github.com/myitcvforks/tools/",
	"github.com/myitcvscratch/tools/",
	"github.com/ningkexin/tools/",
	"github.com/oiooj/tools/",
	"github.com/peterebden/tools/",
	"github.com/phamtanlong/go-crud/",
	"github.com/saibing/",
	"github.com/sauyon/tools/",
	"github.com/tsaikd/tools/",
	"github.com/wxio/tools/",
	"github.com/yousong/tools/",
	"go.coder.com/go-tools/",
}

// Fetcher fetches a list of analyzers.
type Fetcher struct {
	BaseURL    string
	HTTPClient *http.Client
}

func (f *Fetcher) baseURL() string {
	if f.BaseURL == "" {
		return defaultBaseURL
	}
	return f.BaseURL
}

func (f *Fetcher) httpClient() *http.Client {
	if f.HTTPClient == nil {
		return http.DefaultClient
	}
	return f.HTTPClient
}

func (f *Fetcher) List(ctx context.Context) ([]*Info, error) {

	url := f.baseURL() + "golang.org/x/tools/go/analysis?tab=importedby"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := f.httpClient().Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("go.dev returns error with %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var paths []string
	doc.Find(".ImportedBy-list .Details-indent a").Each(func(i int, s *goquery.Selection) {
		path := s.Text()
		for _, ignore := range ignores {
			if strings.HasPrefix(path, ignore) {
				// ignore
				return
			}
		}
		paths = append(paths, path)
	})

	if len(paths) == 0 {
		return nil, nil
	}

	//paths = paths[:100]

	infos := make([]*Info, len(paths))
	eg, ctx := errgroup.WithContext(ctx)

	for i := range paths {
		i := i
		eg.Go(func() error {
			info, err := f.Fetch(ctx, paths[i])
			if err != nil {
				return err
			}
			infos[i] = info
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	var notnils []*Info
	for _, info := range infos {
		if info != nil {
			notnils = append(notnils, info)
		}
	}

	if len(notnils) == 0 {
		return nil, nil
	}

	return notnils, nil
}

func (f *Fetcher) Fetch(ctx context.Context, path string) (*Info, error) {

	for _, ignore := range ignores {
		if strings.HasPrefix(path, ignore) {
			// ignore
			return nil, nil
		}
	}

	dir, err := ioutil.TempDir("", "vetgen")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir) // ignore error

	gopath := filepath.Join(dir, "gopath")
	if err := os.Mkdir(gopath, 0777); err != nil {
		return nil, err
	}

	gocmd, err := exec.LookPath("go")
	if err != nil {
		return nil, err
	}

	goget := exec.CommandContext(ctx, gocmd, "get", path)
	//goget.Stderr = os.Stderr
	goget.Dir = dir
	goget.Env = append(os.Environ(), "GOPATH="+gopath)
	if err := goget.Run(); err != nil {
		// ignore
		return nil, nil
	}

	mode := packages.NeedImports | packages.NeedSyntax | packages.NeedTypes
	cfg := &packages.Config{
		Mode:    mode,
		Context: ctx,
		Dir:     dir,
		Env:     append(os.Environ(), "GOPATH="+gopath),
	}
	pkgs, err := packages.Load(cfg, path)
	if err != nil {
		return nil, err
	}

	if len(pkgs) <= 0 {
		return nil, nil
	}
	pkg := pkgs[0]

	analysisPkg := pkg.Imports["golang.org/x/tools/go/analysis"]
	if analysisPkg == nil {
		return nil, errors.New("cannot get analysis package")
	}

	objAnalyzer := analysisPkg.Types.Scope().Lookup("Analyzer")
	if objAnalyzer == nil {
		return nil, errors.New("cannot get analysis.Analyzer")
	}
	typAnalyzer := types.NewPointer(objAnalyzer.Type())

	for _, ident := range pkg.Types.Scope().Names() {
		obj := pkg.Types.Scope().Lookup(ident)
		if obj != nil && types.Identical(typAnalyzer, obj.Type()) {
			return &Info{
				Import: path,
				Name:   ident,
			}, nil
		}
	}

	return nil, nil
}
