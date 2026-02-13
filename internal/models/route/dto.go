package route

type OrderField string

const (
	OrderFieldCreatedAt OrderField = "created_at"
	OrderFieldPath      OrderField = "path"
	OrderFieldTargetURL OrderField = "target_url"
)

type GetRequest struct {
	ID   *string `param:"id" json:"-"`
	Path *string `param:"id" json:"-"`
}

type ListRequest struct {
	IDs        []string `query:"ids" json:"-"`
	PluginIDs  []string `query:"pids" json:"-"`
	Paths      []string `query:"paths" json:"-"`
	TargetURLs []string `query:"turls" json:"-"`
	Enabled    *bool    `query:"enabled" json:"-"`

	OrderField     OrderField `query:"of" json:"-"`
	OrderDirection string     `query:"od" json:"-"`

	PageSize  int    `query:"ps" json:"-"`
	PageToken string `query:"pt" json:"-"`
}

type CreateRequest struct {
	Path      string `json:"path"`
	TargetURL string `json:"target_url"`

	IdleConnTimeout       int `json:"idle_conn_timeout"`
	TLSHandshakeTimeout   int `json:"tls_handshake_timeout"`
	ExpectContinueTimeout int `json:"expect_continue_timeout"`

	MaxIdleCons           *int `json:"max_idle_conns,omitempty"`
	MaxIdleConsPerHost    *int `json:"max_idle_conns_per_host,omitempty"`
	MaxConsPerHost        *int `json:"max_conns_per_host,omitempty"`
	ResponseHeaderTimeout *int `json:"response_header_timeout"`
}

type DeleteRequest struct {
	ID string `param:"id" json:"-"`
}
