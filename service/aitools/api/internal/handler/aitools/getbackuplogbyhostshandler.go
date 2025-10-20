package aitools

import (
	"net/http"

	"db_ai/service/aitools/api/internal/logic/aitools"
	"db_ai/service/aitools/api/internal/svc"
	"db_ai/service/aitools/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetbackuplogbyhostsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RemoteCommandReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := aitools.NewGetbackuplogbyhostsLogic(r.Context(), svcCtx)
		resp, err := l.Getbackuplogbyhosts(&req)
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.OkJson(w, resp)
		}
	}
}
