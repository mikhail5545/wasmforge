/*
 * Copyright (c) $today.year.Mikhail Kulik.
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import React, { type ReactNode } from "react"
import { Separator } from "@workspace/ui/components/separator"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@workspace/ui/components/card"
import { Badge } from "@workspace/ui/components/badge"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@workspace/ui/components/table"
import { cn } from "@workspace/ui/lib/utils"

interface DocSectionProps {
  id: string
  title: string
  description: string
  children: React.ReactNode
}

interface InfoCardProps {
  title: string
  children: React.ReactNode
}

interface EndpointProps {
  method: string
  path: string
  children?: React.ReactNode
}

interface PayloadTableProps {
  data: { property: string; type: string; description?: string }[]
  className?: string
}

interface TocCrossPageProps {
  title?: string
  data: { label: string; link: string; sub: string[][] }[]
  className?: string
}

const DocSection: React.FC<DocSectionProps> = ({
  id,
  title,
  description,
  children,
}) => {
  return (
    <section id={id} className={"scroll-mt-20"}>
      <Separator className={"mb-8"} />
      <div className={"flex flex-col gap-5"}>
        <div className={"flex flex-col gap-2"}>
          <h2 className={"text-2xl font-bold tracking-tight"}>{title}</h2>
          <p className={"text-muted-foreground"}>{description}</p>
        </div>
        <div className={"flex flex-col gap-5 text-sm leading-7"}>
          {children}
        </div>
      </div>
    </section>
  )
}

const InfoCard: React.FC<InfoCardProps> = ({ title, children }) => {
  return (
    <Card size={"sm"}>
      <CardHeader>
        <CardTitle>{title}</CardTitle>
      </CardHeader>
      <CardContent className={"text-muted-foreground"}>{children}</CardContent>
    </Card>
  )
}

const Endpoint: React.FC<EndpointProps> = ({ method, path, children }) => {
  return (
    <div className={"rounded-lg border bg-card px-3 py-3"}>
      <div className={"flex flex-wrap items-center gap-2"}>
        <Badge variant={"secondary"}>{method}</Badge>
        <code className={"font-mono text-sm"}>{path}</code>
      </div>
      {children && (
        <div className={"mt-2 text-sm leading-6 text-muted-foreground"}>
          {children}
        </div>
      )}
    </div>
  )
}

const PayloadTable: React.FC<PayloadTableProps> = ({ data, className }) => {
  return (
    <div className={cn("overflow-hidden rounded-lg border", className)}>
      <Table>
        <TableHeader className={"sticky top-0 z-10 bg-muted"}>
          <TableRow>
            <TableHead>Property</TableHead>
            <TableHead>Type</TableHead>
            <TableHead>Description</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {data.map((value) => (
            <TableRow key={`${value.property}-${value.type}`}>
              <TableCell>{value.property}</TableCell>
              <TableCell>{value.type}</TableCell>
              <TableCell>{value.description}</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}

const TocCrossPage: React.FC<TocCrossPageProps> = ({
  title,
  data,
  className,
}) => {
  return (
    <div className={cn("flex flex-col gap-3", className)}>
      <p className={"text-2xl font-semibold"}>{title ?? "Table of Contents"}</p>
      <ol className={"mx-5 list-disc space-y-2"}>
        {data.map((item) => (
          <li className={"ml-3"} key={item.link}>
            <a href={item.link} className={"underline"}>
              {item.label}
            </a>
            {item.sub.length > 0 && (
              <ol className={"mx-5 list-disc space-y-1 text-sm"}>
                {item.sub.map(([link, label]) => (
                  <li className={"ml-3"} key={link}>
                    <a href={`${item.link}#${link}`} className={"underline"}>
                      {label}
                    </a>
                  </li>
                ))}
              </ol>
            )}
          </li>
        ))}
      </ol>
    </div>
  )
}

const TocInPage = ({
  title,
  description,
  data,
  className,
}: {
  title?: string
  description?: string
  data: string[][]
  className?: string
}) => {
  return (
    <Card className={className}>
      <CardHeader>
        <CardTitle>{title ?? "Table of Contents"}</CardTitle>
        {description && <CardDescription>{description}</CardDescription>}
      </CardHeader>
      <CardContent>
        <nav className={"grid gap-2 sm:grid-cols-2 lg:grid-cols-4"}>
          {data.map(([id, label]) => (
            <a
              key={id}
              href={`#${id}`}
              className={
                "rounded-lg border px-3 py-2 text-sm font-medium text-muted-foreground transition-colors hover:text-foreground"
              }
            >
              {label}
            </a>
          ))}
        </nav>
      </CardContent>
    </Card>
  )
}

function ReferenceSection({
  title,
  description,
  children,
}: {
  title: string
  description: string
  children: ReactNode
}) {
  return (
    <section className={"scroll-mt-20"}>
      <Separator className={"mb-8"} />
      <div className={"flex flex-col gap-5"}>
        <div className={"flex flex-col gap-2"}>
          <h2 className={"text-2xl font-bold tracking-tight"}>{title}</h2>
          <p className={"text-muted-foreground"}>{description}</p>
        </div>
        <div className={"flex flex-col gap-5"}>{children}</div>
      </div>
    </section>
  )
}

function EndpointGrid({ children }: { children: ReactNode }) {
  return <div className={"grid gap-3"}>{children}</div>
}

export {
  DocSection,
  InfoCard,
  Endpoint,
  PayloadTable,
  TocCrossPage,
  TocInPage,
  ReferenceSection,
  EndpointGrid,
}
