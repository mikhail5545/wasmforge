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
import {motion} from "motion/react";
import {ChevronRight, MoveRight} from "lucide-react";

export default function RoutesPage() {
    const links = [
        { label: "Dashboard", href: "/dashboard" },
        { label: "Users", href: "/users" },
        { label: "Settings", href: "/settings" },
    ];

    const routePaginatedData = usePaginatedData<WasmForge.Route>(
        "/api/routes",
        "routes",
        10,
        "created_at",
        "desc",
        { preload: true }
    );

    const router = useRouter();

    return (
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
                    title={routePaginatedData.error?.message ? "Error creating route" : ""}
                    message={routePaginatedData.error ? routePaginatedData.error.message : ""}
                    isOpen={!!routePaginatedData.error}
                    onClose={() => routePaginatedData.refetch()}
                />
                <div className={"flex flex-col lg:flex-row py-10 px-5 md:px-15 lg:px-30 border-b border-dashed border-stone-800"}>
                    <div className={"border-box w-full lg:w-1/2"}>
                        <h2 className={"text-2xl font-semibold"}>Routes</h2>
                        <p className={"text-sm text-stone-500 mt-5"}>Here you can modify, create and delete routes. All changes will be saved after application restart.</p>
                    </div>
                    <div className={"border-box w-full lg:w-1/2 h-full"}>
                        <div className={"flex flex-col gap-2 items-center lg:items-end"}>
                            <motion.button
                                whileHover={{ scale: 1.15 }}
                                whileTap={{ scale: 0.85 }}
                                className={"px-4 py-2 bg-stone-800 rounded text-sm"}
                                onClick={() => router.push("/routes/new")}
                            >
                                Add new route
                            </motion.button>
                        </div>
                    </div>
                </div>
                <div className={"px-5 md:px-15 lg:px-30 py-10"}>
                    {routePaginatedData.loading ? (
                        <div className={"flex justify-center items-center py-20"}>
                            <div className={"w-10 h-10 border-4 border-t-white border-stone-600 rounded-full animate-spin"}/>
                        </div>
                    ) : (
                        <>
                            {routePaginatedData.data.length === 0 ? (
                                <div className={"w-full pb-10 flex items-center justify-center text-stone-500"}>
                                    <h2 className={"font-lg font-semibold"}>No active routes</h2>
                                </div>
                            ) : (
                                <div className={"grid grid-cols-1 gap-2"}>
                                    {routePaginatedData.data.map((route, idx) => (
                                        <motion.div
                                            key={route.id}
                                            initial={{ opacity: 0, y: 10 }}
                                            animate={{ opacity: 1, y: 0 }}
                                            transition={{ duration: 0.3, delay: idx * 0.1 }}
                                            className={"col-span-1 border border-stone-800 rounded-lg"}
                                        >
                                            <div className={"flex flex-row justify-between"}>
                                                <div className={"flex flex-row gap-4 p-4"}>
                                                    <div className={"flex flex-col gap-1"}>
                                                        <p className={"text-sm text-stone-500"}>Path</p>
                                                        <p className={"text-md font-semibold"}>
                                                            {route.path}
                                                            <MoveRight className={"inline-block ml-4"} size={15}/>
                                                        </p>
                                                    </div>
                                                    <div className={"flex flex-col gap-1"}>
                                                        <p className={"text-sm text-stone-500"}>Target URL</p>
                                                        <p className={"text-md font-semibold"}>{route.target_url}</p>
                                                    </div>
                                                </div>
                                                <div className={"flex max-w-1/7 items-center justify-center"}>
                                                    <motion.a
                                                        className={"w-full h-full flex items-center justify-center bg-stone-800 rounded-r-lg text-sm px-3 hover:bg-stone-700 transition-colors duration-200"}
                                                        href={`/routes/route?path=${route.path}`}
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
                                </div>
                            )}
                        </>
                    )}
                </div>
            </div>
        </div>
    );
}