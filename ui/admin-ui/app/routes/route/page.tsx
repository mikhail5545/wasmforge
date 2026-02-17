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

import {useRouter, useSearchParams} from "next/navigation";
import {useData} from "@/hooks/useData";
import NavBar from "@/components/navigation/NavBar";
import {ErrorDialog} from "@/components/dialog/ErrorDialog";
import {
    CalendarClock,
    Link2,
    Route,
    Eye,
    EyeOff,
    Plus,
    Trash,
    Power,
    Wifi,
    WifiOff,
    X, ChevronLeft, ChevronRight
} from "lucide-react";
import React, {Suspense, useState} from "react";
import {AnimatePresence,motion} from "motion/react";
import {usePaginatedData} from "@/hooks/usePaginatedData";
import {PluginGridListCard, RoutePluginGridListCard} from "@/components/card/GridListCard";

function RoutePageContent() {
    const params = useSearchParams();
    const path = typeof params.get("path") === "string" ? params.get("path")! : "";
    const encoded = encodeURIComponent(path);

    const routeData = useData<WasmForge.Route>(`http://localhost:8080/api/routes/${encoded}`, "route");
    const paginatedRoutePluginsData = usePaginatedData<WasmForge.RoutePlugin>(
        `http://localhost:8080/api/route-plugins?r_ids=${routeData.data?.id}`,
        "route_plugins",
        10,
        "created_at",
        "desc",
        { preload: true },
    );

    const paginatedPluginsData = usePaginatedData<WasmForge.Plugin>(
        `http://localhost:8080/api/plugins`,
        "plugins",
        5,
        "created_at",
        "desc",
        { preload: false },
    );

    const router = useRouter();

    const [showAdvanced, setShowAdvanced] = useState(false);
    const [showPluginSelection, setShowPluginSelection] = useState(false);
    const [selectedPlugin, setSelectedPlugin] = useState<WasmForge.Plugin | null>(null);
    const [showDeleteConfirmation, setShowDeleteConfirmation] = useState(false);

    const links = [
        { label: "Dashboard", href: "/dashboard" },
        { label: "Routes", href: "/routes" },
        { label: "Plugins", href: "/plugins" },
    ];

    return (
        <div className={"flex flex-col w-full"}>
            <NavBar
                title={"WasmForge"}
                links={links}
            />
            <ErrorDialog
                title={routeData.error?.message ? "Error creating route" : ""}
                message={routeData.error ? routeData.error.message : ""}
                isOpen={!!routeData.error}
                onClose={() => routeData.refetch()}
            />
            <div className={"py-10 px-5 md:px-15 lg:px-30"}>
                <div className={"flex w-full lg:w-1/3"}>
                    <AnimatePresence
                        mode={"wait"}
                    >
                        {showPluginSelection && (
                            <motion.div
                                initial={{ opacity: 0 }}
                                animate={{ opacity: 1 }}
                                exit={{ opacity: 0 }}
                                className={"backdrop-blur-2xl fixed inset-0 flex items-start justify-center z-50"}
                            >
                                <motion.div
                                    key={"route-selection"}
                                    initial={{ opacity: 0, y: -20 }}
                                    animate={{ opacity: 1, y: 0 }}
                                    exit={{ opacity: 0, y: -20 }}
                                    transition={{ duration: 0.3 }}
                                    className={"fixed top-20 left-1/2 transform -translate-x-1/2 w-full md:w-1/2 lg:w-1/3 bg-stone-900/80 rounded-md z-50 p-5"}
                                >
                                    <div className={"flex flex-col gap-5"}>
                                        <div className={"flex flex-row items-center justify-start gap-2"}>
                                            <motion.button
                                                whileHover={{ scale: 1.05 }}
                                                whileTap={{ scale: 0.95 }}
                                                onClick={() => {setShowPluginSelection(false); setSelectedPlugin(null);}}
                                                aria-label={"Close plugin selection"}
                                                className={"bg-stone-800 text-sm font-semibold px-2 py-2 rounded disabled:opacity-50 flex items-center justify-center gap-2 hover:bg-red-700/80 transition-colors duration-200"}
                                            >
                                                <X size={12}/>
                                            </motion.button>
                                            <p className={"text-lg font-semibold"}>Select a plugin</p>
                                        </div>
                                        {paginatedPluginsData.loading ? (
                                            <div className={"flex justify-center items-center py-20"}>
                                                <div className={"w-10 h-10 border-4 border-t-white border-stone-600 rounded-full animate-spin"}/>
                                            </div>
                                        ) : (
                                            <div>
                                                {paginatedPluginsData.data.length === 0 ? (
                                                    <div className={"flex flex-col justify-center items-center py-20"}>
                                                        <p className={"text-center text-stone-400"}>No plugins found.</p>
                                                        <a className={"text-center text-stone-400 underline"} href={"/plugins/new"}>Create one.</a>
                                                    </div>
                                                ) : (
                                                    <div className={"flex flex-col gap-5"}>
                                                        <div className={"grid grid-cols-1 gap-1"}>
                                                            {paginatedPluginsData.data.map((plugin, idx) => (
                                                                <PluginGridListCard
                                                                    key={plugin.id}
                                                                    plugin={plugin}
                                                                    index={idx}
                                                                    onClick={() => setSelectedPlugin(plugin)}
                                                                    currentlySelected={selectedPlugin?.id === plugin.id}
                                                                />
                                                            ))}
                                                            <div className={"flex flex-row gap-5 mt-5 justify-between"}>
                                                                <motion.button
                                                                    whileHover={{ scale: 1.05 }}
                                                                    whileTap={{ scale: 0.95 }}
                                                                    onClick={ async () => await paginatedPluginsData.refetch() }
                                                                    className={"border border-stone-800 text-sm font-semibold px-3 py-1 rounded disabled:opacity-50 flex items-center justify-center gap-2"}
                                                                >
                                                                    <ChevronLeft size={10}/>First page
                                                                </motion.button>
                                                                <motion.button
                                                                    whileHover={{ scale: 1.05 }}
                                                                    whileTap={{ scale: 0.95 }}
                                                                    disabled={paginatedPluginsData.nextPageToken === ""}
                                                                    onClick={ async () => { await paginatedPluginsData.nextPage(paginatedPluginsData.nextPageToken, { append: false })} }
                                                                    className={"border border-stone-800 text-sm font-semibold px-3 py-1 rounded disabled:cursor-not-allowed disabled:opacity-50 flex items-center justify-center gap-2"}
                                                                >
                                                                    Next page<ChevronRight size={10}/>
                                                                </motion.button>
                                                            </div>
                                                            <div className={"flex flex-row gap-5 mt-5 justify-between"}>
                                                                <motion.div
                                                                    whileHover={{ scale: 1.05 }}
                                                                    whileTap={{ scale: 0.95 }}
                                                                    onClick={() => router.push(`/routes/plugins/new?route_id=${routeData.data?.id}`)}
                                                                    className={"bg-stone-800 text-sm font-semibold px-2 py-2 rounded disabled:opacity-50 flex items-center justify-center gap-2"}
                                                                >
                                                                    Continue without plugin
                                                                </motion.div>
                                                                <motion.button
                                                                    whileHover={{ scale: 1.05 }}
                                                                    whileTap={{ scale: 0.95 }}
                                                                    disabled={!selectedPlugin}
                                                                    onClick={() => router.push(`/routes/plugins/new?route_id=${routeData.data?.id}&plugin_id=${selectedPlugin?.id}`)}
                                                                    className={"bg-stone-800 text-sm font-semibold px-2 py-2 rounded disabled:opacity-50 flex items-center justify-center gap-2"}
                                                                >
                                                                    Submit selection
                                                                </motion.button>
                                                            </div>
                                                        </div>
                                                    </div>
                                                )}
                                            </div>
                                        )}
                                    </div>
                                </motion.div>
                            </motion.div>
                        )}
                    </AnimatePresence>
                </div>
                {routeData.loading ? (
                    <div className={"flex justify-center items-center py-20"}>
                        <div className={"w-10 h-10 border-4 border-t-white border-stone-600 rounded-full animate-spin"}/>
                    </div>
                ) : (
                    <>
                        <div className={"flex flex-row"}>
                            <a className={"text-lg text-stone-500 underline"} href={"/routes"}>Routes</a>
                            <h2 className={"text-xl"}>{routeData.data?.path}</h2>
                        </div>
                        <div className={"flex flex-col lg:flex-row gap-5 mt-10"}>
                            <div className={"flex w-full lg:w-1/2"}>
                                <div className={"flex flex-col w-full"}>
                                    <div className={"flex flex-row pb-5 items-center justify-start gap-2"}>
                                        <p className={"text-lg font-semibold"}>Associated plugins</p>
                                        <div className={"flex items-center justify-center"}>
                                            <motion.button
                                                whileHover={{ scale: 1.05 }}
                                                whileTap={{ scale: 0.95 }}
                                                onClick={async () => { await paginatedPluginsData.refetch(); setShowPluginSelection(true)}}
                                                className={"bg-stone-800 text-sm font-semibold px-2 py-2 rounded disabled:opacity-50 flex items-center justify-center gap-2"}
                                            >
                                                <Plus size={15}/>
                                            </motion.button>
                                        </div>
                                    </div>
                                    {paginatedRoutePluginsData.loading ? (
                                        <div className={"flex justify-center items-center py-20"}>
                                            <div className={"w-10 h-10 border-4 border-t-white border-stone-600 rounded-full animate-spin"}/>
                                        </div>
                                    ) : (
                                        <div className={"w-full"}>
                                            {paginatedRoutePluginsData.data.length === 0 ? (
                                                <div className={"flex justify-center items-center py-20"}>
                                                    <p className={"text-center text-stone-400"}>No plugins found.</p>
                                                </div>
                                            ) : (
                                                <div>
                                                    <div className={"grid grid-cols-1 gap-1 w-full"}>
                                                        {paginatedRoutePluginsData.data.map((plugin, idx) => (
                                                            <RoutePluginGridListCard
                                                                key={plugin.id}
                                                                routePlugin={plugin}
                                                                index={idx}
                                                                onClick={() => router.push(`/routes/plugins/plugin?id=${plugin.id}`)}
                                                            />
                                                        ))}
                                                    </div>
                                                    <div className={"flex flex-row gap-5 mt-5 justify-between"}>
                                                        <motion.button
                                                            whileHover={{ scale: 1.05 }}
                                                            whileTap={{ scale: 0.95 }}
                                                            onClick={ async () => await paginatedRoutePluginsData.refetch() }
                                                            className={"border border-stone-800 text-sm font-semibold px-3 py-1 rounded disabled:opacity-50 flex items-center justify-center gap-2"}
                                                        >
                                                            <ChevronLeft size={10}/>First page
                                                        </motion.button>
                                                        <motion.button
                                                            whileHover={{ scale: 1.05 }}
                                                            whileTap={{ scale: 0.95 }}
                                                            disabled={paginatedRoutePluginsData.nextPageToken === ""}
                                                            onClick={ async () => { await paginatedRoutePluginsData.nextPage(paginatedRoutePluginsData.nextPageToken, { append: false })} }
                                                            className={"border border-stone-800 text-sm font-semibold px-3 py-1 rounded disabled:cursor-not-allowed disabled:opacity-50 flex items-center justify-center gap-2"}
                                                        >
                                                            Next page<ChevronRight size={10}/>
                                                        </motion.button>
                                                    </div>
                                                </div>
                                            )}
                                        </div>
                                    )}
                                </div>
                            </div>
                            <div className={"flex w-full lg:w-1/2"}>
                                <div className={"w-full"}>
                                    <div className={"flex flex-row pb-5 items-center justify-start gap-2"}>
                                        <p className={"text-lg font-semibold"}>Route details</p>
                                        <div className={"flex items-center justify-center gap-2"}>
                                            <motion.button
                                                whileHover={{ scale: 1.05 }}
                                                whileTap={{ scale: 0.95 }}
                                                onClick={() => setShowAdvanced(prev => !prev)}
                                                aria-label={routeData.data.enabled ? "Disable route" : "Enable route"}
                                                className={`bg-stone-800 text-sm font-semibold px-2 py-2 rounded disabled:opacity-50 flex items-center justify-center gap-2 
                                                    ${routeData.data.enabled ? "hover:bg-red-700/80" : "hover:bg-green-700/80"} transition-colors duration-200`}
                                            >
                                                <Power size={15}/>
                                            </motion.button>
                                            <motion.button
                                                whileHover={{ scale: 1.05 }}
                                                whileTap={{ scale: 0.95 }}
                                                onClick={() => setShowAdvanced(prev => !prev)}
                                                aria-label={"Delete route"}
                                                className={"bg-stone-800 text-sm font-semibold px-2 py-2 rounded disabled:opacity-50 flex items-center justify-center gap-2 hover:bg-red-700/80 transition-colors duration-200"}
                                            >
                                                <Trash size={15}/>
                                            </motion.button>
                                        </div>
                                    </div>
                                    <div className={"grid grid-cols-1 gap-1 w-full"}>
                                        <div className={"col-span-1 rounded-t-xl rounded-b-md bg-stone-800 py-3 flex flex-row justify-between"}>
                                            <div className={"w-1/7 h-full flex justify-center items-center"}>
                                                <div className={"flex justify-center items-center p-4 rounded-full bg-stone-700"}>
                                                    <Route size={15}/>
                                                </div>
                                            </div>
                                            <div className={"w-6/7 flex flex-col"}>
                                                <p className={"text-md text-stone-400"}>Path</p>
                                                <p className={"text-md font-bold"}>{routeData.data.path}</p>
                                            </div>
                                        </div>
                                        <div className={"col-span-1 rounded-md bg-stone-800 py-3 flex flex-row justify-between"}>
                                            <div className={"w-1/7 h-full flex justify-center items-center"}>
                                                <div className={"flex justify-center items-center p-4 rounded-full bg-stone-700"}>
                                                    <Link2 size={15}/>
                                                </div>
                                            </div>
                                            <div className={"w-6/7 flex flex-col"}>
                                                <p className={"text-md text-stone-400"}>Target URL</p>
                                                <p className={"text-md font-bold"}>{routeData.data.target_url}</p>
                                            </div>
                                        </div>
                                        <div className={"col-span-1 rounded-md bg-stone-800 py-3 flex flex-row justify-between"}>
                                            <div className={"w-1/7 h-full flex justify-center items-center"}>
                                                <div className={"flex justify-center items-center p-4 rounded-full bg-stone-700"}>
                                                    {routeData.data.enabled ? (<Wifi size={15}/>) : (<WifiOff size={15}/>)}
                                                </div>
                                            </div>
                                            <div className={"w-6/7 flex flex-col"}>
                                                <p className={"text-md text-stone-400"}>Status</p>
                                                {routeData.data.enabled ? (
                                                    <p className={"text-md text-green-700/80 font-bold"}>enabled</p>
                                                ) : (
                                                    <p className={"text-md text-red-700/80 font-bold"}>disabled</p>
                                                )}
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
                                                <p className={"text-md font-bold"}>{new Date(routeData.data.created_at).toLocaleString()}</p>
                                            </div>
                                        </div>
                                    </div>
                                    <div className={"flex flex-row justify-between"}>
                                        <p className={"text-lg font-semibold py-5"}>Advanced</p>
                                        <div className={"flex items-center justify-center"}>
                                            <motion.button
                                                whileHover={{ scale: 1.05 }}
                                                whileTap={{ scale: 0.95 }}
                                                onClick={() => setShowAdvanced(prev => !prev)}
                                                className={"bg-stone-800 text-sm font-semibold px-3 py-3 rounded disabled:opacity-50 flex items-center justify-center gap-2"}
                                            >
                                                {showAdvanced ? (
                                                    <EyeOff size={12} />
                                                ) : (
                                                    <Eye size={12} />
                                                )}
                                            </motion.button>
                                        </div>
                                    </div>
                                    <AnimatePresence
                                        mode={"wait"}
                                    >
                                        {showAdvanced && (
                                            <motion.div
                                                key={"show-advanced"}
                                                initial={{ opacity: 0, height: 0 }}
                                                animate={{ opacity: 1, height: "auto" }}
                                                exit={{ opacity: 0, height: 0 }}
                                                transition={{ duration: 0.3 }}
                                                className={"overflow-hidden grid grid-cols-1 gap-1 w-full"}
                                            >
                                                <div className={"col-span-1 rounded-t-xl rounded-b-md bg-stone-800 py-3 flex flex-row justify-between"}>
                                                    <div className={"w-1/7 h-full flex justify-center items-center"}>
                                                        <div className={"flex justify-center items-center p-4 rounded-full bg-stone-700"}>
                                                            <Route size={15}/>
                                                        </div>
                                                    </div>
                                                    <div className={"w-6/7 flex flex-col"}>
                                                        <p className={"text-md text-stone-400"}>Idle connection timeout</p>
                                                        <p className={"text-md font-bold"}>{routeData.data.idle_conn_timeout}</p>
                                                    </div>
                                                </div>
                                                <div className={"col-span-1 rounded-md bg-stone-800 py-3 flex flex-row justify-between"}>
                                                    <div className={"w-1/7 h-full flex justify-center items-center"}>
                                                        <div className={"flex justify-center items-center p-4 rounded-full bg-stone-700"}>
                                                            <Route size={15}/>
                                                        </div>
                                                    </div>
                                                    <div className={"w-6/7 flex flex-col"}>
                                                        <p className={"text-md text-stone-400"}>TLS handshake timeout</p>
                                                        <p className={"text-md font-bold"}>{routeData.data.tls_handshake_timeout}</p>
                                                    </div>
                                                </div>
                                                <div className={"col-span-1 rounded-md bg-stone-800 py-3 flex flex-row justify-between"}>
                                                    <div className={"w-1/7 h-full flex justify-center items-center"}>
                                                        <div className={"flex justify-center items-center p-4 rounded-full bg-stone-700"}>
                                                            <Route size={15}/>
                                                        </div>
                                                    </div>
                                                    <div className={"w-6/7 flex flex-col"}>
                                                        <p className={"text-md text-stone-400"}>Expect continue timeout</p>
                                                        <p className={"text-md font-bold"}>{routeData.data.expect_continue_timeout}</p>
                                                    </div>
                                                </div>
                                                <div className={"col-span-1 rounded-md bg-stone-800 py-3 flex flex-row justify-between"}>
                                                    <div className={"w-1/7 h-full flex justify-center items-center"}>
                                                        <div className={"flex justify-center items-center p-4 rounded-full bg-stone-700"}>
                                                            <Route size={15}/>
                                                        </div>
                                                    </div>
                                                    <div className={"w-6/7 flex flex-col"}>
                                                        <p className={"text-md text-stone-400"}>Max idle connections</p>
                                                        <p className={"text-md font-bold"}>{routeData.data.max_idle_cons || "N/A"}</p>
                                                    </div>
                                                </div>
                                                <div className={"col-span-1 rounded-md bg-stone-800 py-3 flex flex-row justify-between"}>
                                                    <div className={"w-1/7 h-full flex justify-center items-center"}>
                                                        <div className={"flex justify-center items-center p-4 rounded-full bg-stone-700"}>
                                                            <Route size={15}/>
                                                        </div>
                                                    </div>
                                                    <div className={"w-6/7 flex flex-col"}>
                                                        <p className={"text-md text-stone-400"}>Max idle connections per host</p>
                                                        <p className={"text-md font-bold"}>{routeData.data.max_idle_cons_per_host || "N/A"}</p>
                                                    </div>
                                                </div>
                                                <div className={"col-span-1 rounded-t-md rounded-b-xl bg-stone-800 py-3 flex flex-row justify-between"}>
                                                    <div className={"w-1/7 h-full flex justify-center items-center"}>
                                                        <div className={"flex justify-center items-center p-4 rounded-full bg-stone-700"}>
                                                            <Route size={15}/>
                                                        </div>
                                                    </div>
                                                    <div className={"w-6/7 flex flex-col"}>
                                                        <p className={"text-md text-stone-400"}>Max connections per host</p>
                                                        <p className={"text-md font-bold"}>{routeData.data.max_cons_per_host || "N/A"}</p>
                                                    </div>
                                                </div>
                                            </motion.div>
                                        )}
                                    </AnimatePresence>
                                </div>
                            </div>
                        </div>
                    </>
                )}
            </div>
        </div>
    );
}

export default function RoutePage() {
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
                <RoutePageContent />
            </div>
        </Suspense>
    );
}