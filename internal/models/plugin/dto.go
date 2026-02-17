/*
 * Copyright (c) 2026. Mikhail Kulik.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package plugin

type OrderField string

const (
	OrderFieldName      OrderField = "name"
	OrderFieldFilename  OrderField = "filename"
	OrderFieldCreatedAt OrderField = "created_at"
)

type GetRequest struct {
	ID       *string `param:"id" json:"-"`
	Name     *string `param:"id" json:"-"`
	Filename *string `param:"id" json:"-"`
}

type ListRequest struct {
	IDs []string `query:"ids" json:"-"`

	Names     []string `query:"n" json:"-"`
	Filenames []string `query:"fn" json:"-"`

	OrderField     OrderField `query:"of" json:"-"`
	OrderDirection string     `query:"od" json:"-"`

	PageSize  int    `query:"ps" json:"-"`
	PageToken string `query:"pt" json:"-"`
}

type CreateRequest struct {
	Name     string `json:"name"`
	Filename string `json:"filename"`
	Data     string `json:"data"` // Base64-encoded plugin data
}

type IDRequest struct {
	ID string `param:"id" json:"-"`
}
