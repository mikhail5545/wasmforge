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
import NavBar from "@/components/navigation/NavBar";
import React, {Suspense, useState, useCallback} from "react";
import {useData} from "@/hooks/useData";
import {ErrorDialog} from "@/components/dialog/ErrorDialog";
import {
    BookKey,
    CalendarClock,
    Lock,
    File,
    ChevronLeft, Trash, X, ChevronRight,
} from "lucide-react";
import {AnimatePresence, motion} from "motion/react";
import {usePaginatedData} from "@/hooks/usePaginatedData";
import {RouteGridListCard} from "@/components/card/GridListCard";
import {useMutation} from "@/hooks/useMutation";
import {ConfirmationDialog} from "@/components/dialog/ConfirmationDialog";

function PluginPageContent() {
    const links = [
        { label: "Dashboard", href: "/dashboard" },
        { label: "Routes", href: "/routes" },
        { label: "Plugins", href: "/plugins" },
    ];

    const router = useRouter();

    const searchParams = useSearchParams();
    const pluginName = typeof searchParams.get("name") === "string" ? searchParams.get("name") : "Unknown Plugin";

    const [selectedRoute, setSelectedRoute] = useState<WasmForge.Route | null>(null);
    const [showRouteSelection, setShowRouteSelection] = useState(false);
    const [showDeleteConfirmation, setShowDeleteConfirmation] = useState(false);

    const pluginData = useData<WasmForge.Plugin>(`http://localhost:8080/api/plugins/${pluginName}`, "plugin");
    const paginatedRoutesData = usePaginatedData<WasmForge.Route>(
        "api/routes",
        "routes",
        10,
        "created_at",
        "desc",
        { preload: true }
    );

    const mutation = useMutation();

    const handleDelete = useCallback(
        async() => {
            if (!pluginData.data) {
                return;
            }
            const response = await mutation.mutate(`https://localhost:8080/api/plugins/${pluginData.data.id}`, "DELETE");
            if (response.success) {
                router.push("/plugins");
            }
        }, [pluginData.data, mutation, router]
    );

    const [globalError, setGlobalError] = useState(pluginData.error || mutation.error || paginatedRoutesData.error);

    return (
        <div className={"flex flex-col w-full"}>
            <NavBar
                title={"WasmForge"}
                links={links}
            />
            <ErrorDialog
                title={globalError?.message || "Error fetching plugin data"}
                message={globalError?.details ||"An unknown error occurred while fetching plugin data."}
                isOpen={!!globalError}
                onClose={() => setGlobalError(null)}
            />
            <ConfirmationDialog
                title={"Delete plugin"}
                message={"Are you sure you want to delete this plugin? This action cannot be undone."}
                isOpen={showDeleteConfirmation}
                accentColor={"red-500"}
                onConfirm={async () => { setShowDeleteConfirmation(false); await handleDelete(); }}
                onCancel={() => setShowDeleteConfirmation(false)}
            />
            <div className={"py-10 px-5 md:px-15 lg:px-30"}>
                <div className={"flex w-full lg:w-1/3"}>
                    <AnimatePresence mode={"wait"}>
                        {showRouteSelection && (
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
                                                onClick={() => {setShowRouteSelection(false); setSelectedRoute(null);}}
                                                aria-label={"Close plugin selection"}
                                                className={"bg-stone-800 text-sm font-semibold px-2 py-2 rounded disabled:opacity-50 flex items-center justify-center gap-2 hover:bg-red-700/80 transition-colors duration-200"}
                                            >
                                                <X size={12}/>
                                            </motion.button>
                                            <p className={"text-lg font-semibold"}>Select a route</p>
                                        </div>
                                        {paginatedRoutesData.loading ? (
                                            <div className={"flex justify-center items-center py-20"}>
                                                <div className={"w-10 h-10 border-4 border-t-white border-stone-600 rounded-full animate-spin"}/>
                                            </div>
                                        ) : (
                                            <div>
                                                {paginatedRoutesData.data.length === 0 ? (
                                                    <div className={"flex flex-col justify-center items-center py-20"}>
                                                        <p className={"text-center text-stone-400"}>No routes found.</p>
                                                        <a className={"text-center text-stone-400 underline"} href={"/routes/new"}>Create one.</a>
                                                    </div>
                                                ) : (
                                                    <div className={"flex flex-col gap-5"}>
                                                        <div className={"grid grid-cols-1 gap-1"}>
                                                            {paginatedRoutesData.data.map((route, idx) => (
                                                                <RouteGridListCard
                                                                    key={route.id}
                                                                    route={route}
                                                                    index={idx}
                                                                    onClick={() => setSelectedRoute(route)}
                                                                    currentlySelected={selectedRoute?.id === route.id}
                                                                />
                                                            ))}
                                                        </div>
                                                        <div className={"flex flex-row gap-5 mt-5 justify-between"}>
                                                            <motion.button
                                                                whileHover={{ scale: 1.05 }}
                                                                whileTap={{ scale: 0.95 }}
                                                                onClick={ async () => await paginatedRoutesData.refetch() }
                                                                className={"border border-stone-800 text-sm font-semibold px-3 py-1 rounded disabled:opacity-50 flex items-center justify-center gap-2"}
                                                            >
                                                                <ChevronLeft size={10}/>First page
                                                            </motion.button>
                                                            <motion.button
                                                                whileHover={{ scale: 1.05 }}
                                                                whileTap={{ scale: 0.95 }}
                                                                disabled={paginatedRoutesData.nextPageToken === ""}
                                                                onClick={ async () => { await paginatedRoutesData.nextPage(paginatedRoutesData.nextPageToken, { append: false })} }
                                                                className={"border border-stone-800 text-sm font-semibold px-3 py-1 rounded disabled:cursor-not-allowed disabled:opacity-50 flex items-center justify-center gap-2"}
                                                            >
                                                                Next page<ChevronRight size={10}/>
                                                            </motion.button>
                                                        </div>
                                                        <div className={"flex flex-row gap-5 mt-5 justify-between"}>
                                                            <motion.div
                                                                whileHover={{ scale: 1.05 }}
                                                                whileTap={{ scale: 0.95 }}
                                                                onClick={() => router.push(`/routes/plugins/new?plugin_id=${pluginData.data?.id}`)}
                                                                className={"bg-stone-800 text-sm font-semibold px-2 py-2 rounded disabled:opacity-50 flex items-center justify-center gap-2"}
                                                            >
                                                                Continue without route
                                                            </motion.div>
                                                            <motion.button
                                                                whileHover={{ scale: 1.05 }}
                                                                whileTap={{ scale: 0.95 }}
                                                                disabled={!selectedRoute}
                                                                onClick={() => router.push(`/routes/plugins/new?plugin_id=${pluginData.data?.id}&route_id=${selectedRoute?.id}`)}
                                                                className={"bg-stone-800 text-sm font-semibold px-2 py-2 rounded disabled:opacity-50 flex items-center justify-center gap-2"}
                                                            >
                                                                Submit selection
                                                            </motion.button>
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
                {pluginData.loading ? (
                    <div className={"flex justify-center items-center py-20"}>
                        <div className={"w-10 h-10 border-4 border-t-white border-stone-600 rounded-full animate-spin"}/>
                    </div>
                ) : (
                    <>
                        <div className={"flex flex-row items-center justify-between gap-2 mb-5"}>
                            <h1 className={"text-3xl font-bold"}>Plugin Details</h1>
                        </div>
                        <div className={"flex flex-row"}>
                            <a className={"text-lg text-stone-500 underline"} href={"/plugins"}>Plugins</a>
                            <h2 className={"text-xl"}>{`/${pluginData.data?.name}`}</h2>
                        </div>
                        <div className={"flex flex-col lg:flex-row gap-5"}>
                            <div className={"flex w-full lg:w-2/3"}>
                                <div className={"flex w-full flex-col p-5 rounded"}>
                                    <p className={"text-xl font-semibold"}>Details</p>
                                    <div className={"grid grid-cols-1 gap-1 w-full mt-5"}>
                                        <div className={"col-span-1 rounded-t-xl rounded-b-md bg-stone-800 py-3 flex flex-row justify-between"}>
                                            <div className={"w-1/7 h-full flex justify-center items-center"}>
                                                <div className={"flex justify-center items-center p-4 rounded-full bg-stone-700"}>
                                                    <BookKey size={15}/>
                                                </div>
                                            </div>
                                            <div className={"w-6/7 flex flex-col"}>
                                                <p className={"text-md text-stone-400"}>Name</p>
                                                <p className={"text-md font-bold"}>{pluginData.data.name}</p>
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
                                                <p className={"text-md font-bold"}>{pluginData.data.filename}</p>
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
                                                <p className={"text-md font-bold truncate"}>{`SHA-256:${pluginData.data.checksum}`}</p>
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
                                                <p className={"text-md font-bold"}>{new Date(pluginData.data.created_at).toLocaleString()}</p>
                                            </div>
                                        </div>
                                    </div>
                                </div>
                            </div>
                            <div className={"flex w-full lg:w-1/3"}>
                                <div className={"flex w-full flex-col p-5 rounded"}>
                                    <p className={"text-xl font-semibold"}>Actions</p>
                                    <div className={"border-box border border-stone-800 mt-5 rounded"}>
                                        <div className={"flex flex-row items-center justify-center p-5 gap-5"}>
                                            <motion.button
                                                whileHover={{ scale: 1.05 }}
                                                whileTap={{ scale: 0.95 }}
                                                onClick={() => setShowRouteSelection(true)}
                                                className={"bg-stone-800 max-w-1/2 text-sm font-semibold px-3 py-3 rounded disabled:opacity-50 flex items-center justify-center gap-2 hover:bg-stone-700/80 transition-colors duration-200"}
                                            >
                                                Attach to route
                                            </motion.button>
                                            <motion.button
                                                whileHover={{ scale: 1.05 }}
                                                whileTap={{ scale: 0.95 }}
                                                onClick={() => setShowDeleteConfirmation(true)}
                                                className={"bg-stone-800 max-w-1/2 text-sm font-semibold px-3 py-3 rounded disabled:opacity-50 flex items-center justify-center gap-2 hover:bg-red-700/80 transition-colors duration-200"}
                                            >
                                                Delete
                                            </motion.button>
                                        </div>
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

export default function PluginPage() {
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
                <PluginPageContent />
            </div>
        </Suspense>
    );
}