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

import React, {useState} from "react";
import {usePaginatedData} from "@/hooks/usePaginatedData";
import PageLayout from "@/components/layout/PageLayout";
import NavBar from "@/components/navigation/NavBar";
import {motion,AnimatePresence} from "motion/react";
import {Funnel, ListFilter, Grid2X2, List, ChevronRight, ArrowUpRight, Plus} from "lucide-react";

export default function RoutesPage() {
    const links = [
        { label: "Routes", href: "/routes", active: true },
        { label: "Plugins", href: "/plugins", active: false },
        { label: "Settings", href: "/settings", active: false },
    ];

    const routePaginatedData = usePaginatedData<WasmForge.Route>(
        "/api/routes",
        "routes",
        10,
        "created_at",
        "desc",
        { preload: true }
    );

    const [viewMode, setViewMode] = useState<"list" | "grid">("list");

    return (
        <PageLayout>
            <NavBar links={links}/>
            <div className={"flex flex-col gap-5 w-full mt-20"}>
                <div className={"flex flex-col lg:flex-row gap-5"}>
                    <div className={"lg:w-1/4"}>
                        <div className={"flex flex-col w-full items-start p-5 gap-5 bg-stone-800 rounded-4xl"}>
                            <div className={"flex flex-row justify-between items-center w-full"}>
                                <p className={"text-lg font-semibold"}>Your Routes</p>
                                <div className={"flex items-center justify-center gap-2"}>
                                    <motion.a
                                        whileHover={{ scale: 1.05 }}
                                        whileTap={{ scale: 0.95 }}
                                        href={"/routes/new"}
                                        className={"p-2 rounded-4xl bg-white text-black hover:bg-white/80 transition-colors duration-200"}
                                    >
                                        <Plus size={15}/>
                                    </motion.a>
                                    {viewMode === "list" ? (
                                        <motion.button
                                            whileHover={{ scale: 1.05 }}
                                            whileTap={{ scale: 0.95 }}
                                            onClick={() => setViewMode("grid")}
                                            className={"p-2 rounded-4xl bg-white text-black hover:bg-white/80 transition-colors duration-200"}
                                        >
                                            <Grid2X2 size={15}/>
                                        </motion.button>
                                    ) : (
                                        <motion.button
                                            whileHover={{ scale: 1.05 }}
                                            whileTap={{ scale: 0.95 }}
                                            onClick={() => setViewMode("list")}
                                            className={"p-3 rounded-4xl bg-white text-black hover:bg-white/80 transition-colors duration-200"}
                                        >
                                            <List size={15}/>
                                        </motion.button>
                                    )}
                                </div>
                            </div>
                            <div className={"flex flex-col space-y-4 w-full"}>
                                <div className={"flex flex-row items-center gap-4"}>
                                    <Funnel size={15}/>
                                    <p className={"text-md"}>Filters</p>
                                </div>
                                <div className={"flex flex-row items-center gap-4"}>
                                    <ListFilter size={15}/>
                                    <p className={"text-md"}>Sort</p>
                                </div>
                            </div>
                        </div>
                    </div>
                    <div className={"lg:w-3/4"}>
                        <div className={"flex flex-col gap-5 w-full"}>
                            {routePaginatedData.loading ? (
                                <div className={"flex items-center justify-center py-40"}>
                                    <div className={"w-10 h-10 border-4 border-t-white border-stone-600 rounded-full animate-spin"}/>
                                </div>
                            ) : (
                                <div>
                                    {routePaginatedData.data.length === 0 ? (
                                        <div className={"flex flex-col items-center justify-center text-center py-40"}>
                                            <p className={"text-lg"}>You didn't create any routes yet</p>
                                            <a href={"/routes/new"} className={"text-lg underline"}>Start with creating one</a>
                                        </div>
                                    ) : (
                                        <AnimatePresence mode={"wait"}>
                                            {viewMode === "list" && (
                                                <motion.div
                                                    key={"list-view"}
                                                    initial={{ opacity: 0, y: 10 }}
                                                    animate={{ opacity: 1, y: 0 }}
                                                    exit={{ opacity: 0, y: 10 }}
                                                    transition={{ duration: 0.3 }}
                                                    className={"grid grid-cols-1 gap-5 w-full"}
                                                >
                                                    {routePaginatedData.data.map((route, idx) => (
                                                        <motion.div
                                                            initial={{ opacity: 0, y: 10 }}
                                                            animate={{ opacity: 1, y: 0 }}
                                                            exit={{ opacity: 0, y: 10 }}
                                                            transition={{ duration: 0.3, delay: idx * 0.1 }}
                                                            key={route.id}
                                                            className={"col-span-1 flex flex-row justify-between items-center rounded-4xl bg-stone-800 p-4"}
                                                        >
                                                            <div className={"px-3 flex flex-row w-4/5 gap-10"}>
                                                                <div className={"flex flex-col w-1/3"}>
                                                                    <p className={"text-sm"}>Path</p>
                                                                    <p className={"text-md font-bold truncate"}>{route.path}</p>
                                                                </div>
                                                                <div className={"flex flex-col w-1/3"}>
                                                                    <p className={"text-sm"}>Target URL</p>
                                                                    <p className={"text-md font-bold truncate"}>{route.target_url}</p>
                                                                </div>
                                                                <div className={"flex flex-col w-1/3"}>
                                                                    <p className={"text-sm"}>Created at</p>
                                                                    <p className={"text-md font-bold truncate"}>{new Date(route.created_at).toLocaleString()}</p>
                                                                </div>
                                                            </div>
                                                            <div className={"flex items-center justify-center h-full"}>
                                                                <motion.a
                                                                    whileHover={{ scale: 1.05 }}
                                                                    whileTap={{ scale: 0.95 }}
                                                                    href={`/routes/route?path=${route.path}`}
                                                                    className={"h-full px-4 items-center justify-center flex bg-amber-500 hover:bg-amber-500/80 transition-colors duration-200 rounded-2xl"}
                                                                >
                                                                    <ChevronRight size={15}/>
                                                                </motion.a>
                                                            </div>
                                                        </motion.div>
                                                    ))}
                                                </motion.div>
                                            )}
                                            {viewMode === "grid" && (
                                                <motion.div
                                                    key={"grid-view"}
                                                    initial={{ opacity: 0, y: 10 }}
                                                    animate={{ opacity: 1, y: 0 }}
                                                    exit={{ opacity: 0, y: 10 }}
                                                    transition={{ duration: 0.3 }}
                                                    className={"grid grid-cols-1 md:grid-cols-3 gap-5 w-full"}
                                                >
                                                    {routePaginatedData.data.map((route, idx) => (
                                                        <motion.div
                                                            key={route.id}
                                                            initial={{ opacity: 0, y: 10 }}
                                                            animate={{ opacity: 1, y: 0 }}
                                                            exit={{ opacity: 0, y: 10 }}
                                                            transition={{ duration: 0.3, delay: idx * 0.1 }}
                                                            className={"col-span-1 flex flex-col gap-5 rounded-4xl bg-stone-800 p-4"}
                                                        >
                                                            <div className={"flex flex-row justify-between items-center"}>
                                                                <div className={"flex flex-row gap-2 items-center"}>
                                                                    <p className={"text-lg font-semibold"}>{route.path}</p>
                                                                    <div className={"items-center flex justify-center"}>
                                                                        <div className={`h-2 w-2 rounded-full animate-pulse ${route.enabled ? "bg-green-500" : "bg-red-500"}`}/>
                                                                    </div>
                                                                </div>
                                                                <div className={"justify-center items-center flex"}>
                                                                    <motion.a
                                                                        href={`/routes/route?path=${route.path}`}
                                                                        whileHover={{ scale: 1.05 }}
                                                                        whileTap={{ scale: 0.95 }}
                                                                        className={"p-2 rounded-full bg-white text-black"}
                                                                    >
                                                                        <ArrowUpRight size={15}/>
                                                                    </motion.a>
                                                                </div>
                                                            </div>
                                                            <div className={"flex flex-col p-2 rounded-lg bg-stone-900"}>
                                                                    <p className={"text-md"}>Target URL</p>
                                                                    <p className={"text-md font-semibold"}>{route.target_url}</p>
                                                            </div>
                                                            <div className={"flex flex-col p-2 rounded-lg bg-stone-900"}>
                                                                <p className={"text-md"}>Created at</p>
                                                                <p className={"text-md font-semibold"}>{new Date(route.created_at).toLocaleString()}</p>
                                                            </div>
                                                        </motion.div>
                                                    ))}
                                                </motion.div>
                                            )}
                                        </AnimatePresence>
                                    )}
                                </div>
                            )}
                        </div>
                    </div>
                </div>
            </div>
        </PageLayout>
    );
}