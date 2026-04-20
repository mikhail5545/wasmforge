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

'use client';

import React, {useEffect, useMemo} from "react";
import PageLayout from "@/components/layout/PageLayout";
import NavBar from "@/components/navigation/NavBar";
import {useData} from "@/hooks/useData";
import {motion} from "motion/react";

function toSortedPercentages(input: Record<string, number> | undefined): Array<[string, number]> {
    if (!input) {
        return [];
    }
    return Object.entries(input).sort((a, b) => b[1] - a[1]);
}

export default function StatsPage() {
    const links = [
        { label: "Routes", href: "/routes", active: false },
        { label: "Plugins", href: "/plugins", active: false },
        { label: "Stats", href: "/stats", active: true },
        { label: "Settings", href: "/settings", active: false },
    ];

    const overview = useData<WasmForge.ProxyStatsOverview>(
        "http://localhost:8080/api/proxy/stats/overview",
        "overview"
    );
    const routeStats = useData<WasmForge.ProxyRouteStats[]>(
        "http://localhost:8080/api/proxy/stats/routes?limit=10",
        "routes"
    );
    const timeseries = useData<WasmForge.ProxyTimeseriesPoint[]>(
        "http://localhost:8080/api/proxy/stats/timeseries?bucket_seconds=60",
        "timeseries"
    );
    const refetchOverview = overview.refetch;
    const refetchRouteStats = routeStats.refetch;
    const refetchTimeseries = timeseries.refetch;

    useEffect(() => {
        const interval = window.setInterval(() => {
            void refetchOverview();
            void refetchRouteStats();
            void refetchTimeseries();
        }, 5000);
        return () => window.clearInterval(interval);
    }, [refetchOverview, refetchRouteStats, refetchTimeseries]);

    const statusPercentages = useMemo(() => {
        return toSortedPercentages(overview.data?.status_code_percentages);
    }, [overview.data]);

    const maxSeriesCount = useMemo(() => {
        if (!timeseries.data || timeseries.data.length === 0) {
            return 1;
        }
        return Math.max(...timeseries.data.map((point) => point.total_requests), 1);
    }, [timeseries.data]);

    return (
        <PageLayout>
            <NavBar links={links}/>
            <div className={"flex flex-col gap-5 w-full mt-20"}>
                <div className={"grid grid-cols-1 xl:grid-cols-4 gap-5"}>
                    <div className={"col-span-1 bg-stone-800 rounded-4xl p-5"}>
                        <p className={"text-lg font-semibold"}>Total Requests</p>
                        <p className={"text-4xl mt-3"}>{overview.data?.total_requests ?? 0}</p>
                    </div>
                    <div className={"col-span-1 bg-stone-800 rounded-4xl p-5"}>
                        <p className={"text-lg font-semibold"}>Average RPS</p>
                        <p className={"text-4xl mt-3"}>{(overview.data?.avg_rps ?? 0).toFixed(2)}</p>
                    </div>
                    <div className={"col-span-1 bg-stone-800 rounded-4xl p-5"}>
                        <p className={"text-lg font-semibold"}>Average Latency</p>
                        <p className={"text-4xl mt-3"}>{(overview.data?.avg_latency_ms ?? 0).toFixed(2)} ms</p>
                    </div>
                    <div className={"col-span-1 bg-stone-800 rounded-4xl p-5"}>
                        <p className={"text-lg font-semibold"}>Dropped Events</p>
                        <p className={"text-4xl mt-3"}>{overview.data?.dropped_events ?? 0}</p>
                    </div>
                </div>

                <div className={"grid grid-cols-1 xl:grid-cols-3 gap-5"}>
                    <div className={"col-span-1 bg-stone-800 rounded-4xl p-5"}>
                        <p className={"text-xl font-semibold mb-4"}>Status Code Percentages</p>
                        {overview.loading ? (
                            <div className={"flex items-center justify-center py-10"}>
                                <div className={"w-8 h-8 border-4 border-t-white border-stone-600 rounded-full animate-spin"}/>
                            </div>
                        ) : statusPercentages.length === 0 ? (
                            <p className={"opacity-75"}>No data in current window</p>
                        ) : (
                            <div className={"flex flex-col gap-3"}>
                                {statusPercentages.map(([code, pct]) => (
                                    <div key={code} className={"flex flex-row items-center justify-between gap-4"}>
                                        <p className={"font-semibold"}>{code}</p>
                                        <div className={"flex-1 h-2 bg-stone-700 rounded-full overflow-hidden"}>
                                            <div className={"h-full bg-amber-500 rounded-full"} style={{ width: `${Math.min(pct, 100)}%` }}/>
                                        </div>
                                        <p>{pct.toFixed(2)}%</p>
                                    </div>
                                ))}
                            </div>
                        )}
                    </div>

                    <div className={"col-span-1 xl:col-span-2 bg-stone-800 rounded-4xl p-5"}>
                        <p className={"text-xl font-semibold mb-4"}>Top Routes</p>
                        {routeStats.loading ? (
                            <div className={"flex items-center justify-center py-10"}>
                                <div className={"w-8 h-8 border-4 border-t-white border-stone-600 rounded-full animate-spin"}/>
                            </div>
                        ) : !routeStats.data || routeStats.data.length === 0 ? (
                            <p className={"opacity-75"}>No route metrics yet</p>
                        ) : (
                            <div className={"grid grid-cols-1 gap-3"}>
                                {routeStats.data.map((route, idx) => (
                                    <motion.div
                                        key={`${route.route_path}-${idx}`}
                                        initial={{ opacity: 0, y: 8 }}
                                        animate={{ opacity: 1, y: 0 }}
                                        transition={{ duration: 0.2, delay: idx * 0.05 }}
                                        className={"p-4 rounded-3xl bg-stone-700/50 flex flex-col lg:flex-row lg:items-center lg:justify-between gap-3"}
                                    >
                                        <div className={"flex flex-col"}>
                                            <p className={"text-sm opacity-70"}>Route</p>
                                            <p className={"font-semibold break-all"}>{route.route_path}</p>
                                        </div>
                                        <div className={"grid grid-cols-1 sm:grid-cols-3 gap-3 lg:min-w-[420px]"}>
                                            <div>
                                                <p className={"text-sm opacity-70"}>Requests</p>
                                                <p>{route.total_requests}</p>
                                            </div>
                                            <div>
                                                <p className={"text-sm opacity-70"}>Avg RPS</p>
                                                <p>{route.avg_rps.toFixed(2)}</p>
                                            </div>
                                            <div>
                                                <p className={"text-sm opacity-70"}>Avg Latency</p>
                                                <p>{route.avg_latency_ms.toFixed(2)} ms</p>
                                            </div>
                                        </div>
                                    </motion.div>
                                ))}
                            </div>
                        )}
                    </div>
                </div>

                <div className={"bg-stone-800 rounded-4xl p-5"}>
                    <p className={"text-xl font-semibold mb-4"}>Requests Timeline (1-minute buckets)</p>
                    {timeseries.loading ? (
                        <div className={"flex items-center justify-center py-10"}>
                            <div className={"w-8 h-8 border-4 border-t-white border-stone-600 rounded-full animate-spin"}/>
                        </div>
                    ) : !timeseries.data || timeseries.data.length === 0 ? (
                        <p className={"opacity-75"}>No timeseries data yet</p>
                    ) : (
                        <div className={"grid grid-cols-1 gap-3"}>
                            {timeseries.data.slice(-20).map((point) => (
                                <div key={point.bucket_start} className={"grid grid-cols-1 lg:grid-cols-4 gap-3 items-center"}>
                                    <p className={"text-sm opacity-80"}>{new Date(point.bucket_start).toLocaleTimeString()}</p>
                                    <div className={"lg:col-span-2 h-2 bg-stone-700 rounded-full overflow-hidden"}>
                                        <div
                                            className={"h-full bg-amber-500 rounded-full"}
                                            style={{ width: `${(point.total_requests / maxSeriesCount) * 100}%` }}
                                        />
                                    </div>
                                    <div className={"flex flex-row items-center gap-4"}>
                                        <p>{point.total_requests} req</p>
                                        <p className={"opacity-80"}>{point.avg_latency_ms.toFixed(2)} ms</p>
                                    </div>
                                </div>
                            ))}
                        </div>
                    )}
                </div>
            </div>
        </PageLayout>
    );
}
