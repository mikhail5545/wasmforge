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

package plugins

type OrderField string

const (
	OrderFieldCreatedAt      OrderField = "created_at"
	OrderFieldExecutionOrder OrderField = "execution_order"
)

type GetRequest struct {
	ID string `param:"id" json:"-"`
}

type ListRequest struct {
	IDs       []string `query:"ids" json:"-"`
	RouteIDs  []string `query:"r_ids" json:"-"`
	PluginIDs []string `query:"p_ds" json:"-"`

	OrderField     OrderField `query:"of" json:"-"`
	OrderDirection string     `query:"od" json:"-"`

	PageSize  int    `query:"ps" json:"-"`
	PageToken string `query:"pt" json:"-"`
}

type CreateRequest struct {
	RouteID           string `json:"route_id"`
	PluginID          string `json:"plugin_id"`
	VersionConstraint string `json:"version_constraint"`
	ExecutionOrder    int    `json:"execution_order"`

	Config *string `json:"config,omitempty"`
}

type UpdateRequest struct {
	ID string `param:"id" json:"-"`

	PluginID          *string `json:"plugin_id,omitempty"`
	VersionConstraint *string `json:"version_constraint,omitempty"`
	ExecutionOrder    *int    `json:"execution_order,omitempty"`
	Config            *string `json:"config,omitempty"`
}

type DeleteRequest struct {
	ID string `param:"id" json:"-"`
}
