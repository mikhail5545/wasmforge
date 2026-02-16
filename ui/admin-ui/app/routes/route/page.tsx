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

import {useSearchParams} from "next/navigation";
import {useData} from "@/hooks/useData";
import NavBar from "@/components/navigation/NavBar";
import {ErrorDialog} from "@/components/dialog/ErrorDialog";
import {
    Route,
    Server,
    CirclePlay,
    RefreshCwOff,
    ArrowRight,
} from "lucide-react";
import {Suspense, useState} from "react";
import {AnimatePresence,motion} from "motion/react";
export default function RoutePage() {
    const params = useSearchParams();
    const path = typeof params.get("path") === "string" ? params.get("path")! : "";
    const encoded = encodeURIComponent(path);

    const routeData = useData<WasmForge.Route>(`http://localhost:8080/api/routes/${encoded}`, "route");

    const [showTimingsConfig, setShowTimingsConfig] = useState(false);

    const links = [
        { label: "Dashboard", href: "/dashboard" },
        { label: "Users", href: "/users" },
        { label: "Settings", href: "/settings" },
    ];

    if (routeData.loading) {
        return (
            <div className={"flex min-h-screen bg-stone-950 font-mono text-white"}>
                <div className={"flex flex-col w-full"}>
                    <div className={"flex justify-center items-center py-20"}>
                        <div className={"w-10 h-10 border-4 border-t-white border-stone-600 rounded-full animate-spin"}/>
                    </div>
                </div>
            </div>
        );
    }

    return (
        <Suspense>
        <div className={"flex min-h-screen bg-stone-950 font-mono text-white"}>
            <div className={"flex flex-col w-full"}>
                <NavBar
                    title={"Admin UI"}
                    links={links}
                    rightContent={
                        <div className={"flex items-center gap-2"}>
                            <button className={"px-3 py-1 bg-stone-800 rounded text-sm"}>Sign out</button>
                        </div>
                    }
                />
                <ErrorDialog
                    title={routeData.error?.message ? "Error creating route" : ""}
                    message={routeData.error ? routeData.error.message : ""}
                    isOpen={!!routeData.error}
                    onClose={() => routeData.refetch()}
                />
                <div className={"py-10 px-5 md:px-15 lg:px-30"}>
                    <div className={"flex flex-row "}>
                        <a className={"text-lg text-stone-500 underline"} href={"/routes"}>Routes</a>
                        <h2 className={"text-xl"}>{routeData.data.path}</h2>
                    </div>
                    <div className={"flex flex-col lg:flex-row gap-2"}>
                        <div className={"flex h-full w-full lg:w-1/2"}>
                           <div className={"grid grid-cols-1 gap-1 w-full"}>

                           </div>
                        </div>
                        <div className={"flex flex-col h-full w-full lg:w-1/2"}>
                            <div className={"grid grid-cols-1 gap-1 w-full"}>
                                <div className={"col-span-1 rounded-t-xl rounded-b-md bg-stone-800 py-3 flex flex-row justify-between"}>
                                    <div className={"w-1/5 h-full flex justify-center items-center"}>
                                        <Route size={15}/>
                                    </div>
                                    <div className={"w-4/5 flex flex-col"}>
                                        <p className={"text-md text-stone-400"}>Path</p>
                                        <p className={"text-md font-bold"}>{routeData.data.path}</p>
                                    </div>
                                </div>
                                <div className={"col-span-1 rounded-md bg-stone-800 py-3 flex flex-row justify-between"}>
                                    <div className={"w-1/5 h-full flex justify-center items-center"}>
                                        <Server size={15}/>
                                    </div>
                                    <div className={"w-4/5 flex flex-col"}>
                                        <p className={"text-md text-stone-400"}>Target URL</p>
                                        <p className={"text-md font-bold"}>{routeData.data.target_url}</p>
                                    </div>
                                </div>
                                <div className={"col-span-1 rounded-t-md rounded-b-xl bg-stone-800 py-3 flex flex-row justify-between"}>
                                    <div className={"w-1/5 h-full flex justify-center items-center"}>
                                        {routeData.data.enabled ? (
                                            <CirclePlay size={15} color={"#22c55e"}/>
                                        ) : (
                                            <RefreshCwOff size={15} color={"#ef4444"}/>
                                        )}
                                    </div>
                                    <div className={"w-4/5 flex flex-col"}>
                                        <p className={"text-md text-stone-400"}>Status</p>
                                        {routeData.data.enabled ? (
                                            <p className={"text-md font-bold text-green-500"}>Enabled</p>
                                        ) : (
                                            <p className={"text-md font-bold text-red-500"}>Disabled</p>
                                        )}
                                    </div>
                                </div>
                            </div>
                            <div className={"flex flex-row justify-between py-5"}>
                                <p className={"text-lg font-semibold"}>Timing Configuration</p>
                                <motion.button
                                    className={"px-3 py-1 mt-2 bg-stone-800 rounded text-sm"}
                                    onClick={() => setShowTimingsConfig((prev) => !prev)}
                                    whileHover={{ scale: 1.05 }}
                                    whileTap={{ scale: 0.95 }}
                                >
                                    {showTimingsConfig ? (
                                        <p>Hide</p>
                                    ) : (
                                        <p>Show</p>
                                    )}
                                </motion.button>
                            </div>
                            <AnimatePresence>
                                {showTimingsConfig && (
                                    <motion.div
                                        className={"grid grid-cols-1 gap-1"}
                                        key={"modal"}
                                        initial={{ opacity: 0, y: -10 }}
                                        animate={{ opacity: 1, y: 0 }}
                                        exit={{ opacity: 0, y: -10 }}
                                    >
                                        <div className={"col-span-1 rounded-t-xl rounded-b-md bg-stone-800 py-3 flex flex-row justify-between"}>
                                            <div className={"w-1/5 h-full flex justify-center items-center"}>
                                                <Server size={15}/>
                                            </div>
                                            <div className={"w-4/5 flex flex-col"}>
                                                <p className={"text-md text-stone-400"}>Idle connection timeout</p>
                                                <p className={"text-md font-bold"}>{routeData.data.idle_conn_timeout}</p>
                                            </div>
                                        </div>
                                        <div className={"col-span-1 rounded-md bg-stone-800 py-3 flex flex-row justify-between"}>
                                            <div className={"w-1/5 h-full flex justify-center items-center"}>
                                                <Server size={15}/>
                                            </div>
                                            <div className={"w-4/5 flex flex-col"}>
                                                <p className={"text-md text-stone-400"}>TLS handshake timeout</p>
                                                <p className={"text-md font-bold"}>{routeData.data.tls_handshake_timeout}</p>
                                            </div>
                                        </div>
                                        <div className={"col-span-1 rounded-md bg-stone-800 py-3 flex flex-row justify-between"}>
                                            <div className={"w-1/5 h-full flex justify-center items-center"}>
                                                <Server size={15}/>
                                            </div>
                                            <div className={"w-4/5 flex flex-col"}>
                                                <p className={"text-md text-stone-400"}>Expect continue timeout</p>
                                                <p className={"text-md font-bold"}>{routeData.data.expect_continue_timeout}</p>
                                            </div>
                                        </div>
                                        <div className={"col-span-1 rounded-md bg-stone-800 py-3 flex flex-row justify-between"}>
                                            <div className={"w-1/5 h-full flex justify-center items-center"}>
                                                <Server size={15}/>
                                            </div>
                                            <div className={"w-4/5 flex flex-col"}>
                                                <p className={"text-md text-stone-400"}>Max idle connections per host</p>
                                                <p className={"text-md font-bold"}>{routeData.data.max_idle_cons_per_host || "N/A"}</p>
                                            </div>
                                        </div>
                                        <div className={"col-span-1 rounded-md bg-stone-800 py-3 flex flex-row justify-between"}>
                                            <div className={"w-1/5 h-full flex justify-center items-center"}>
                                                <Server size={15}/>
                                            </div>
                                            <div className={"w-4/5 flex flex-col"}>
                                                <p className={"text-md text-stone-400"}>Max idle connections</p>
                                                <p className={"text-md font-bold"}>{routeData.data.max_idle_cons || "N/A"}</p>
                                            </div>
                                        </div>
                                        <div className={"col-span-1 rounded-t-md rounded-b-xl bg-stone-800 py-3 flex flex-row justify-between"}>
                                            <div className={"w-1/5 h-full flex justify-center items-center"}>
                                                <Server size={15}/>
                                            </div>
                                            <div className={"w-4/5 flex flex-col"}>
                                                <p className={"text-md text-stone-400"}>Response header timeout</p>
                                                <p className={"text-md font-bold"}>{routeData.data.response_header_timeout || "N/A"}</p>
                                            </div>
                                        </div>
                                    </motion.div>
                                )}
                            </AnimatePresence>
                        </div>
                    </div>
                </div>
            </div>
        </div>
        </Suspense>
    );
}