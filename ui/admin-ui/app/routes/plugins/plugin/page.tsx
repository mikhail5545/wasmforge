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

import React, {Suspense} from "react";
import {useRouter, useSearchParams} from "next/navigation";
import {useData} from "@/hooks/useData";
import {ErrorDialog} from "@/components/dialog/ErrorDialog";
import NavBar from "@/components/navigation/NavBar";
import {
    BookKey,
    CalendarClock,
    Lock,
    File,
    Shuffle, ChevronRight, Trash,
} from "lucide-react";
import {motion} from "motion/react";
import {githubDarkTheme, JsonEditor} from "json-edit-react";

function RoutePluginPageContent() {
    const links = [
        { label: "Dashboard", href: "/dashboard" },
        { label: "Routes", href: "/routes" },
        { label: "Plugins", href: "/plugins" },
    ];

    const searchParams = useSearchParams();
    const router = useRouter();

    const pluginId = searchParams.get("id") || "Undefined";
    const routePluginData = useData<WasmForge.RoutePlugin>(`http://localhost:8080/api/route-plugins/${pluginId}`, "route_plugin");

    return (
        <div className={"flex flex-col w-full"}>
            <NavBar
                title={"WasmForge"}
                links={links}
            />
            <ErrorDialog
                title={routePluginData.error?.message || "Unexpected error occurred"}
                message={routePluginData.error?.details || "No additional details available. Try to reload page"}
                isOpen={!!routePluginData.error}
                onClose={() => routePluginData.refetch()}
            />
            <div className={"py-10 px-5 md:px-15 lg:px-30"}>
                {routePluginData.loading ? (
                    <div className={"flex justify-center items-center py-20"}>
                        <div className={"w-10 h-10 border-4 border-t-white border-stone-600 rounded-full animate-spin"}/>
                    </div>
                ) : (
                    <>
                        <div className={"flex flex-row items-center justify-between gap-2 mb-5"}>
                            <h1 className={"text-3xl font-bold"}>Route Plugin Details</h1>
                            <div className={"flex flex-row h-full gap-4"}>
                                <motion.button
                                    whileHover={{ scale: 1.05 }}
                                    whileTap={{ scale: 0.95 }}
                                    onClick={() => router.push(`/routes/route?path=${routePluginData.data.route_id}`)}
                                    className={"bg-stone-800 text-sm font-semibold px-3 py-2 rounded flex items-center justify-center gap-2 hover:bg-stone-700/80 transition-colors duration-200"}
                                >
                                    Route details
                                </motion.button>
                                <motion.button
                                    whileHover={{ scale: 1.05 }}
                                    whileTap={{ scale: 0.95 }}
                                    className={"bg-stone-800 text-sm font-semibold px-3 py-3 rounded disabled:opacity-50 flex items-center justify-center gap-2 hover:bg-red-700/80 transition-colors duration-200"}
                                >
                                    <Trash size={15}/>
                                </motion.button>
                            </div>
                        </div>
                        <div className={"flex flex-col lg:flex-row gap-5"}>
                            <div className={"flex w-full lg:w-1/2"}>
                                <div className={"flex w-full flex-col p-5 rounded"}>
                                    <p className={"text-xl font-semibold"}>Essentials</p>
                                    <div className={"grid grid-cols-1 gap-1 mt-3"}>
                                        <div className={"col-span-1 rounded-t-xl rounded-b-md bg-stone-800 py-3 flex flex-row justify-between"}>
                                            <div className={"w-1/7 h-full flex justify-center items-center"}>
                                                <div className={"flex justify-center items-center p-4 rounded-full bg-stone-700"}>
                                                    <Shuffle size={15}/>
                                                </div>
                                            </div>
                                            <div className={"w-6/7 flex flex-col"}>
                                                <p className={"text-md text-stone-400"}>Execution order</p>
                                                <p className={"text-md font-bold"}>{routePluginData.data.execution_order}</p>
                                            </div>
                                        </div>
                                        <div className={"col-span-1 rounded-t-md rounded-b-xl bg-stone-800 py-3 flex flex-row justify-between"}>
                                            <div className={"w-1/7 h-full flex justify-center items-center"}>
                                                <div className={"flex justify-center items-center p-4 rounded-full bg-stone-700"}>
                                                    <CalendarClock size={15}/>
                                                </div>
                                            </div>
                                            <div className={"w-6/7 flex flex-col"}>
                                                <p className={"text-md text-stone-400"}>Created at</p>
                                                <p className={"text-md font-bold"}>{new Date(routePluginData.data.created_at).toLocaleString()}</p>
                                            </div>
                                        </div>
                                    </div>
                                    <div className={"flex flex-row items-center justify-between mt-5"}>
                                        <p className={"text-xl font-semibold"}>Plugin details</p>
                                        <div className={"flex justify-center items-center h-full"}>
                                            <motion.button
                                                whileHover={{ scale: 1.05 }}
                                                whileTap={{ scale: 0.95 }}
                                                className={"bg-stone-800 text-sm font-semibold px-2 py-2 rounded disabled:opacity-50 flex items-center justify-center gap-2"}
                                                onClick={() => router.push(`/plugins/plugin?name=${routePluginData.data.plugin?.name}`)}
                                            >
                                                <ChevronRight size={15}/>
                                            </motion.button>
                                        </div>
                                    </div>
                                    {routePluginData.data.plugin ? (
                                        <div className={"grid grid-cols-1 gap-1 mt-3"}>
                                            <div className={"col-span-1 rounded-t-xl rounded-b-md bg-stone-800 py-3 flex flex-row justify-between"}>
                                                <div className={"w-1/7 h-full flex justify-center items-center"}>
                                                    <div className={"flex justify-center items-center p-4 rounded-full bg-stone-700"}>
                                                        <BookKey size={15}/>
                                                    </div>
                                                </div>
                                                <div className={"w-6/7 flex flex-col"}>
                                                    <p className={"text-md text-stone-400"}>Name</p>
                                                    <p className={"text-md font-bold"}>{routePluginData.data.plugin.name}</p>
                                                </div>
                                            </div>
                                            <div className={"col-span-1 rounded-md bg-stone-800 py-3 flex flex-row justify-between"}>
                                                <div className={"w-1/7 h-full flex justify-center items-center"}>
                                                    <div className={"flex justify-center items-center p-4 rounded-full bg-stone-700"}>
                                                        <File size={15}/>
                                                    </div>
                                                </div>
                                                <div className={"w-6/7 flex flex-col"}>
                                                    <p className={"text-md text-stone-400"}>Filename</p>
                                                    <p className={"text-md font-bold"}>{routePluginData.data.plugin.filename}</p>
                                                </div>
                                            </div>
                                            <div className={"col-span-1 rounded-md bg-stone-800 py-3 flex flex-row justify-between"}>
                                                <div className={"w-1/7 h-full flex justify-center items-center"}>
                                                    <div className={"flex justify-center items-center p-4 rounded-full bg-stone-700"}>
                                                        <Lock size={15}/>
                                                    </div>
                                                </div>
                                                <div className={"w-6/7 flex flex-col"}>
                                                    <p className={"text-md text-stone-400"}>Checksum</p>
                                                    <p className={"text-md font-bold truncate"}>{`SHA-256:${routePluginData.data.plugin.checksum}`}</p>
                                                </div>
                                            </div>
                                            <div className={"col-span-1 rounded-t-md rounded-b-xl bg-stone-800 py-3 flex flex-row justify-between"}>
                                                <div className={"w-1/7 h-full flex justify-center items-center"}>
                                                    <div className={"flex justify-center items-center p-4 rounded-full bg-stone-700"}>
                                                        <CalendarClock size={15}/>
                                                    </div>
                                                </div>
                                                <div className={"w-6/7 flex flex-col"}>
                                                    <p className={"text-md text-stone-400"}>Created at</p>
                                                    <p className={"text-md font-bold"}>{new Date(routePluginData.data.plugin.created_at).toLocaleString()}</p>
                                                </div>
                                            </div>
                                        </div>
                                    ) : (
                                        <div className={"flex flex-col items-center justify-center py-10"}>
                                            <p className={"text-md font-semibold"}>Plugin details are unavailable</p>
                                            <p className={"text-sm font-semibold text-stone-400"}>Try reloading page</p>
                                        </div>
                                    )}
                                </div>
                            </div>
                            <div className={"flex w-full lg:w-1/2"}>
                                <div className={"flex w-full flex-col p-5 rounded"}>
                                    <p className={"text-xl font-semibold"}>Configuration</p>
                                    <div className={"border-box rounded bg-stone-800 mt-5 p-10"}>
                                        <JsonEditor data={JSON.parse(routePluginData.data.config)} theme={githubDarkTheme}/>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </>
                )}
            </div>
        </div>
    );
}

export default function RoutePluginPage() {
    return (
        <Suspense fallback={
            <div className={"flex min-h-screen bg-stone-950 font-mono text-white"}>
                <div className={"flex flex-col w-full"}>
                    <div className={"flex justify-center items-center py-20"}>
                        <div className={"w-10 h-10 border-4 border-t-white border-stone-600 rounded-full animate-spin"}/>
                    </div>
                </div>
            </div>
        }>
            <div className={"flex min-h-screen bg-stone-950 font-mono text-white"}>
                <RoutePluginPageContent />
            </div>
        </Suspense>
    );
}