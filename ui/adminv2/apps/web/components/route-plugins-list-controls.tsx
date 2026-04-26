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
  Ellipsis,
  Grid2X2,
  Plus,
  Sheet,
} from "lucide-react"
import { Button } from "@workspace/ui/components/button"

interface RoutePluginsListControlsProps {
  orderField: string
  setOrderField: React.Dispatch<React.SetStateAction<string>>
  orderDirection: string
  setOrderDirection: React.Dispatch<React.SetStateAction<string>>
  className?: string
  viewMode?: string
  setViewMode?: React.Dispatch<React.SetStateAction<string>>
  showCreateButton: boolean
  createUrlOverride?: string
}

const RoutePluginsListControls: React.FC<RoutePluginsListControlsProps> = ({
  orderField,
  setOrderField,
  orderDirection,
  setOrderDirection,
  className,
  viewMode,
  setViewMode,
  showCreateButton,
  createUrlOverride,
}) => {
 return (
   <div className={cn("flex flex-row items-center gap-2", className)}>
     {showCreateButton && (
       <Button variant={"outline"} asChild>
         <a href={createUrlOverride ?? "/routes/plugins/new"}>
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
         {showCreateButton && (
           <DropdownMenuItem asChild>
             <a href={createUrlOverride ?? "/routes/plugins/new"}>
               <Plus />
               New
             </a>
           </DropdownMenuItem>
         )}
         <DropdownMenuSub>
           <DropdownMenuSubTrigger inset>Order by</DropdownMenuSubTrigger>
           <DropdownMenuSubContent>
             <DropdownMenuRadioGroup
               value={orderField}
               onValueChange={setOrderField}
             >
               <DropdownMenuRadioItem value={"execution_order"}>
                 Execution Order
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
       </DropdownMenuContent>
     </DropdownMenu>
   </div>
 )
}

export { RoutePluginsListControls }