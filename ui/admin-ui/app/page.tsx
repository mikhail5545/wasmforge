'use client';

import NavBar from "@/components/navigation/NavBar";
import React from "react";
import {motion} from "motion/react";
import {
    MoveRight,
    ChevronRight,
} from "lucide-react";

import {usePaginatedData} from "@/hooks/usePaginatedData";
import {ErrorDialog} from "@/components/dialog/ErrorDialog";

export default function Home() {
    const links = [
        { label: "Dashboard", href: "/dashboard" },
        { label: "Users", href: "/users" },
        { label: "Settings", href: "/settings" },
    ];

    const routes: WasmForge.Route[] = [
        { id: "4a6d9b8f-b034-48be-b559-71e7b3fb9cbf", createdAt: "31.12.1992", path: "/user", enabled: true, targetUrl: "http://localhost:8080/user", idleConnTimeout: 0, tlsHandshakeTimeout: 0, expectedContinueTimeout: 0 },
        { id: "0003d4f4-370e-47ae-94c1-b812feeb7f71", createdAt: "04.10.2018", path: "/admin", enabled: true, targetUrl: "http://localhost:8080/admin", idleConnTimeout: 0, tlsHandshakeTimeout: 0, expectedContinueTimeout: 0 },
    ];

    const routePaginatedData = usePaginatedData<WasmForge.Route>(
        "/api/routes?e=true",
        "routes",
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
                    rightContent={
                        <div className={"flex items-center gap-2"}>
                            <button className={"px-3 py-1 bg-stone-800 rounded text-sm"}>Sign out</button>
                        </div>
                    }
                />
                <ErrorDialog
                    title={routePaginatedData.error ? "Error fetching routes" : ""}
                    message={routePaginatedData.error ? routePaginatedData.error.message : ""}
                    isOpen={!!routePaginatedData.error}
                    onClose={() => routePaginatedData.refetch()}
                />
                <div className={"flex flex-col xl:flex-row py-10 px-5 md:px-15 lg:px-30"}>
                    <div className={"border-box w-full lg:w-1/2"}>
                        <h2 className={"text-2xl font-semibold"}>Control Panel</h2>
                        <p className={""}></p>
                    </div>
                </div>
                <div className={"px-5 md:px-15 lg:px-30 py-10"}>
                    <div className={"flex flex-col xl:flex-row gap-4"}>
                        <div className={"border-box w-full xl:w-1/2"}>
                            <div className={"w-full rounded-4xl border border-stone-800 py-5 px-7"}>
                                <div className={"flex flex-col gap-2"}>
                                    <div className={"flex flex-row justify-between"}>
                                        <div className={"flex flex-col items-start gap-1"}>
                                            <div className={"flex flex-row items-center justify-center gap-4"}>
                                                <h2 className={"text-xl font-semibold"}>Proxy Server</h2>
                                                <motion.div
                                                    animate={{ opacity: 0.5 }}
                                                    transition={{ duration: 0.7, repeat: Infinity, repeatType: "reverse" }}
                                                    className={"w-3 h-3 rounded-full bg-green-500/70"}
                                                    />
                                                </div>
                                                <p className={"text-stone-400 font-semibold text-sm"}>Listening on :9000</p>
                                            </div>
                                        <div className={"flex flex-col items-start gap-1"}>
                                            <button className={"px-3 py-1 bg-stone-800 rounded text-sm w-full"}>Restart</button>
                                            <button className={"px-3 py-1 bg-red-700/70 rounded text-sm w-full"}>Stop</button>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                        <div className={"w-full xl:w-1/2 rounded-4xl border border-stone-800 py-5 px-7"}>
                            <div className={"p-2 pb-5"}>
                                <a className={"text-xl font-semibold underline"} href={"/routes"}>Routes<ChevronRight size={25} className={"inline-block ml-3"}/></a>
                            </div>
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
                                        ) :  (
                                            <div className={"grid grid-cols-1 gap-4"}>
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
                                                                    <p className={"text-sm text-stone-500"}>Middleware</p>
                                                                    <p className={"text-md font-semibold"}>3 plugin(s)<MoveRight className={"inline-block ml-4"} size={15}/></p>
                                                                </div>
                                                                <div className={"flex flex-col gap-1"}>
                                                                    <p className={"text-sm text-stone-500"}>Target URL</p>
                                                                    <p className={"text-md font-semibold"}>{route.targetUrl}</p>
                                                                </div>
                                                            </div>
                                                            <div className={"flex max-w-1/7 items-center justify-center"}>
                                                                <motion.a
                                                                    className={"w-full h-full flex items-center justify-center bg-stone-800 rounded-r-lg text-sm px-3 hover:bg-stone-700 transition-colors duration-200"}
                                                                    href={`/routes/${route.path.replace(/\//g, "_")}`}
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
                                                <div className={"w-full items-center text-center"}>
                                                    <a className={"text-sm font-semibold text-stone-500 underline"} href={"/routes"}>All routes</a>
                                                </div>
                                            </div>
                                        )}
                                </>
                            )}
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
}