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

import NavBar from "@/components/navigation/NavBar";
import React from "react";
import {useRouter} from "next/navigation";
import {usePaginatedData} from "@/hooks/usePaginatedData";
import {ErrorDialog} from "@/components/dialog/ErrorDialog";
import {AnimatePresence, motion} from "motion/react";
import {
    ChevronRight,
    MoveRight,
    Grid2X2,
    List
} from "lucide-react";

export default function PluginsPage() {
    const links = [
        { label: "Dashboard", href: "/dashboard" },
        { label: "Routes", href: "/routes" },
        { label: "Plugins", href: "/plugins" },
    ];

    const router = useRouter();
    const [viewMode, setViewMode] = React.useState<"list" | "grid">("list");

    const pluginPaginatedData = usePaginatedData<WasmForge.Plugin>(
        "/api/plugins",
        "plugins",
        10,
        "created_at",
        "desc",
        { preload: true }
    );

    return (
        <div className={"flex min-h-screen bg-stone-950 font-mono text-white"}>
            <div className={"flex flex-col w-full"}>
                <NavBar
                    title={"Admin UI"}
                    links={links}
                />
                <ErrorDialog
                    title={pluginPaginatedData.error?.message ? "Error retrieving plugins" : ""}
                    message={pluginPaginatedData.error ? pluginPaginatedData.error.message : ""}
                    isOpen={!!pluginPaginatedData.error}
                    onClose={() => pluginPaginatedData.refetch()}
                />
                <div className={"flex flex-col lg:flex-row py-10 px-5 md:px-15 lg:px-30 border-b border-dashed border-stone-800"}>
                    <div className={"border-box w-full lg:w-1/2"}>
                        <h2 className={"text-2xl font-semibold"}>Plugins</h2>
                        <p className={"text-sm text-stone-500 mt-5"}>Here you can modify, create, upload and delete plugins. All changes will be saved after application restart.</p>
                    </div>
                    <div className={"border-box w-full lg:w-1/2 h-full"}>
                        <div className={"flex flex-col gap-2 items-center lg:items-end"}>
                            <motion.button
                                whileHover={{ scale: 1.05 }}
                                whileTap={{ scale: 0.95 }}
                                className={"px-4 py-2 bg-stone-800 rounded text-sm"}
                                onClick={() => router.push("/plugins/new")}
                            >
                                Add new plugin
                            </motion.button>
                        </div>
                    </div>
                </div>
                <div className={"flex flex-row justify-between py-7 px-5 md:px-15 lg:px-30 border-b border-dashed border-stone-800 w-full max-h-10"}>
                    <div className={"flex items-center gap-2"}>
                        <motion.button
                            whileHover={{ scale: 1.05 }}
                            whileTap={{ scale: 0.95 }}
                            onClick={() => setViewMode(viewMode === "list" ? "grid" : "list")}
                            className={"px-4 py-2 bg-white text-black rounded-xl text-sm hover:bg-black hover:text-white transition-colors duration-200 border border-white"}
                        >
                            <div className={"w-full h-full flex items-center justify-center gap-1 text-center"}>
                                {viewMode === "list" ? (
                                    <p>Grid<Grid2X2 size={12} className={"inline-block ml-2"}/></p>
                                ) : (
                                    <p>List<List size={12} className={"inline-block ml-2"}/></p>
                                )}
                            </div>
                        </motion.button>
                    </div>
                </div>
                <div className={"px-5 md:px-15 lg:px-30 py-10"}>
                    {pluginPaginatedData.loading ? (
                        <div className={"flex justify-center items-center py-20"}>
                            <div className={"w-10 h-10 border-4 border-t-white border-stone-600 rounded-full animate-spin"}/>
                        </div>
                    ) : (
                        <>
                            {pluginPaginatedData.data.length === 0 ? (
                                <div className={"w-full pb-10 flex items-center justify-center text-stone-500"}>
                                    <h2 className={"font-lg font-semibold"}>You didn't create any plugins yet</h2>
                                </div>
                            ) : (
                                <AnimatePresence
                                    mode={"wait"}
                                >
                                    {viewMode === "list" && (
                                        <motion.div
                                            key={"list"}
                                            initial={{ opacity: 0, x: -20 }}
                                            animate={{ opacity: 1, x: 0 }}
                                            exit={{ opacity: 0, x: 20 }}
                                            transition={{ duration: 0.3 }}
                                            className={"grid grid-cols-1 gap-2"}
                                        >
                                            {pluginPaginatedData.data.map((plugin, idx) => (
                                                <motion.div
                                                    key={plugin.id}
                                                    initial={{ opacity: 0, y: 10 }}
                                                    animate={{ opacity: 1, y: 0 }}
                                                    transition={{ duration: 0.3, delay: idx * 0.1 }}
                                                    className={"col-span-1 border border-stone-800 rounded-lg"}
                                                >
                                                    <div className={"flex flex-row justify-between p-5"}>
                                                        <div className={"flex flex-row gap-5 w-2/3"}>
                                                            <div className={"flex flex-col gap-2 w-1/2"}>
                                                                <p className={"text-sm text-stone-500"}>Name</p>
                                                                <p className={"text-md font-semibold"}>{plugin.name}</p>
                                                            </div>
                                                            <div className={"flex flex-col gap-2 w-1/2"}>
                                                                <p className={"text-sm text-stone-500"}>Filename</p>
                                                                <p className={"text-md font-semibold"}>{plugin.filename}</p>
                                                            </div>
                                                        </div>
                                                        <div className={"flex max-w-1/3 items-center justify-center"}>
                                                            <motion.a
                                                                className={"flex items-center justify-center p-3 rounded-lg bg-stone-800 hover:bg-stone-700 transition-colors duration-200"}
                                                                href={`/plugins/plugin?name=${plugin.name}`}
                                                                whileHover={{ scale: 1.05 }}
                                                                whileTap={{ scale: 0.95 }}
                                                                transition={{ duration: 0.3, delay: 0.1 }}
                                                            >
                                                                <ChevronRight size={25}/>
                                                            </motion.a>
                                                        </div>
                                                    </div>
                                                </motion.div>
                                            ))}
                                        </motion.div>
                                    )}
                                    {viewMode === "grid" && (
                                        <motion.div
                                            key={"grid"}
                                            initial={{ opacity: 0, x: -20 }}
                                            animate={{ opacity: 1, x: 0 }}
                                            exit={{ opacity: 0, x: 20 }}
                                            transition={{ duration: 0.3 }}
                                            className={"grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-2 md:gap-3 lg:gap-5"}
                                        >
                                            {pluginPaginatedData.data.map((plugin, idx) => (
                                                <motion.div
                                                    key={plugin.id}
                                                    initial={{ opacity: 0, y: 10 }}
                                                    animate={{ opacity: 1, y: 0 }}
                                                    transition={{ duration: 0.3, delay: idx * 0.1 }}
                                                    className={"col-span-1 border border-stone-800 rounded-lg"}
                                                >
                                                    <div className={"flex flex-col space-y-2 p-5"}>
                                                        <div className={"flex flex-row justify-between"}>
                                                            <p className={"text-md text-stone-400"}>Name</p>
                                                            <p className={"text-md font-semibold truncate max-w-2/3"}>{plugin.name}</p>
                                                        </div>
                                                        <div className={"flex flex-row justify-between"}>
                                                            <p className={"text-md text-stone-400"}>Filename</p>
                                                            <p className={"text-md font-semibold truncate max-w-2/3"}>{plugin.filename}</p>
                                                        </div>
                                                        <div className={"flex flex-row justify-between"}>
                                                            <p className={"text-md text-stone-400"}>Created at</p>
                                                            <p className={"text-md font-semibold"}>{new Date(plugin.created_at).toLocaleString()}</p>
                                                        </div>
                                                        <motion.a
                                                            className={"mt-5 flex max-h-5 py-5 items-center justify-center w-full p-3 bg-white text-black rounded-lg hover:bg-stone-800 hover:text-white transition-colors duration-200"}
                                                            href={`/plugins/plugin?name=${plugin.name}`}
                                                            whileHover={{ scale: 1.05 }}
                                                            whileTap={{ scale: 0.95 }}
                                                            transition={{ duration: 0.3, delay: 0.1 }}
                                                        >
                                                            Details <ChevronRight size={15} className={"inline-block ml-2"}/>
                                                        </motion.a>
                                                    </div>
                                                </motion.div>
                                            ))}
                                        </motion.div>
                                    )}
                                </AnimatePresence>
                            )}
                        </>
                    )}
                </div>
            </div>
        </div>
    );
}