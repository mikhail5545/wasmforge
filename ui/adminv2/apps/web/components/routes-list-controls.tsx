/*
 * Copyright (c) 2026. Mikhail Kulik.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import React from "react"
import { cn } from "@workspace/ui/lib/utils"
import { Field, FieldGroup } from "@workspace/ui/components/field"
import {
  InputGroup,
  InputGroupAddon,
  InputGroupButton,
  InputGroupInput,
} from "@workspace/ui/components/input-group"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuSeparator,
  DropdownMenuSub,
  DropdownMenuSubContent,
  DropdownMenuSubTrigger,
  DropdownMenuTrigger,
} from "@workspace/ui/components/dropdown-menu"
import {
  ChevronDown,
  Ellipsis,
  Grid2X2,
  Plus,
  Search,
  Sheet,
} from "lucide-react"
import { Button } from "@workspace/ui/components/button"
import { PaginatedData } from "@/hooks/use-paginated-data"
import { Route } from "types/route"

interface RoutesListControlsProps {
  searchDisabled?: boolean
  orderField: string
  setOrderField: React.Dispatch<React.SetStateAction<string>>
  orderDirection: string
  setOrderDirection: React.Dispatch<React.SetStateAction<string>>
  className?: string
  viewMode?: string
  setViewMode?: React.Dispatch<React.SetStateAction<string>>
  showCreateButton: boolean
  routesData: PaginatedData<Route>
  createUrlOverride?: string
}

const RoutesListControls: React.FC<RoutesListControlsProps> = ({
  searchDisabled,
  orderField,
  setOrderField,
  orderDirection,
  setOrderDirection,
  className,
  viewMode,
  setViewMode,
  showCreateButton,
  routesData,
  createUrlOverride,
}) => {
  const [searchQuery, setSearchQuery] = React.useState("")
  const [searchBy, setSearchBy] = React.useState("path")

  const refetchWithQuery = React.useCallback(async () => {
    const trimmedQuery = searchQuery.trim()

    if (trimmedQuery === "") {
      routesData.setQueryParams({})
      await routesData.refetch()
      return
    }

    routesData.setQueryParams(
      searchBy === "path" ? { paths: trimmedQuery } : { turls: trimmedQuery }
    )

    await routesData.refetch()
  }, [routesData, searchBy, searchQuery])

  return (
    <div className={cn("flex flex-row items-center gap-2", className)}>
      <FieldGroup className={"flex max-w-sm flex-row"}>
        <Field>
          <InputGroup>
            <InputGroupInput
              aria-label={"search plugins"}
              type={"text"}
              value={searchQuery}
              placeholder={
                searchBy === "path"
                  ? "/api/resource"
                  : "https://example.com/api/resource"
              }
              onChange={(e) => setSearchQuery(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter") {
                  void refetchWithQuery()
                }
              }}
            />
            <InputGroupAddon className={"block-start"}>
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <InputGroupButton>
                    <span>Search by</span>
                    <ChevronDown className={"mt-1 inline-block"} />
                  </InputGroupButton>
                </DropdownMenuTrigger>
                <DropdownMenuContent>
                  <DropdownMenuRadioGroup
                    value={searchBy}
                    onValueChange={setSearchBy}
                  >
                    <DropdownMenuRadioItem value={"path"}>
                      Path
                    </DropdownMenuRadioItem>
                    <DropdownMenuRadioItem value={"target_url"}>
                      Target URL
                    </DropdownMenuRadioItem>
                  </DropdownMenuRadioGroup>
                </DropdownMenuContent>
              </DropdownMenu>
            </InputGroupAddon>
          </InputGroup>
        </Field>
        <Button
          variant={"outline"}
          size={"icon"}
          onClick={() => {
            void refetchWithQuery()
          }}
          disabled={searchDisabled}
        >
          <Search />
        </Button>
      </FieldGroup>
      <div className={"flex flex-row gap-2"}>
        {showCreateButton && (
          <Button variant={"outline"} asChild>
            <a href={createUrlOverride || "/routes/new"}>
              <Plus />
              New
            </a>
          </Button>
        )}
        {viewMode && setViewMode && (
          <Button
            variant={"outline"}
            size={"icon"}
            onClick={() => setViewMode(viewMode === "table" ? "grid" : "table")}
          >
            {viewMode === "table" ? <Grid2X2 /> : <Sheet />}
          </Button>
        )}
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant={"outline"} size={"icon"}>
              <Ellipsis />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent>
            <DropdownMenuItem asChild>
              <a href={"/routes/new"}>
                <Plus />
                New
              </a>
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuSub>
              <DropdownMenuSubTrigger inset>Order by</DropdownMenuSubTrigger>
              <DropdownMenuSubContent>
                <DropdownMenuRadioGroup
                  value={orderField}
                  onValueChange={setOrderField}
                >
                  <DropdownMenuRadioItem value={"path"}>
                    Path
                  </DropdownMenuRadioItem>
                  <DropdownMenuRadioItem value={"target_url"}>
                    Target URL
                  </DropdownMenuRadioItem>
                  <DropdownMenuRadioItem value={"created_at"}>
                    Created at
                  </DropdownMenuRadioItem>
                </DropdownMenuRadioGroup>
              </DropdownMenuSubContent>
            </DropdownMenuSub>
            <DropdownMenuSub>
              <DropdownMenuSubTrigger inset>Direction</DropdownMenuSubTrigger>
              <DropdownMenuSubContent>
                <DropdownMenuRadioGroup
                  value={orderDirection}
                  onValueChange={setOrderDirection}
                >
                  <DropdownMenuRadioItem value={"asc"}>
                    Ascending
                  </DropdownMenuRadioItem>
                  <DropdownMenuRadioItem value={"desc"}>
                    Descending
                  </DropdownMenuRadioItem>
                </DropdownMenuRadioGroup>
              </DropdownMenuSubContent>
            </DropdownMenuSub>
            {viewMode && setViewMode && (
              <DropdownMenuSub>
                <DropdownMenuSubTrigger inset>View</DropdownMenuSubTrigger>
                <DropdownMenuSubContent>
                  <DropdownMenuRadioGroup
                    value={viewMode}
                    onValueChange={setViewMode}
                  >
                    <DropdownMenuRadioItem value={"table"}>
                      Table
                    </DropdownMenuRadioItem>
                    <DropdownMenuRadioItem value={"grid"}>
                      Grid
                    </DropdownMenuRadioItem>
                  </DropdownMenuRadioGroup>
                </DropdownMenuSubContent>
              </DropdownMenuSub>
            )}
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </div>
  )
}

export { RoutesListControls }
