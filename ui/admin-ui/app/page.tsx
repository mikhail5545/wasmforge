'use client';

import NavBar from "@/components/navigation/NavBar";
import React from "react";
import PageLayout from "@/components/layout/PageLayout";
import {motion,AnimatePresence} from "motion/react";
import {usePaginatedData} from "@/hooks/usePaginatedData";
import {ArrowUpRight, ChevronRight, Power} from "lucide-react";
import {useRouter} from "next/navigation";
import {useData} from "@/hooks/useData";

export default function Home() {
    const routePaginatedData = usePaginatedData<WasmForge.Route>(
        "/api/routes?enabled=true",
        "routes",
        10,
        "created_at",
        "desc",
        { preload: true }
    );
    const pluginPaginatedData = usePaginatedData<WasmForge.Plugin>(
        "/api/plugins",
        "plugins",
        5,
        "created_at",
        "desc",
        { preload: true }
    );
    const proxyServerStatus = useData<WasmForge.ProxyServerStatus>("http://localhost:8080/api/proxy/config", "status");

    const router = useRouter();
    const links = [
        { label: "Routes", href: "/routes", active: false },
        { label: "Plugins", href: "/plugins", active: false },
        { label: "Settings", href: "/settings", active: false },
    ];

    return (
        <PageLayout>
            <NavBar links={links}/>
            <div className={"flex flex-col lg:flex-row gap-5 w-full mt-20"}>
                <div className={"border-box lg:w-1/5"}>
                    <div className={"border-box  bg-stone-800 rounded-4xl p-5"}>
                        {proxyServerStatus.loading ?
                            (
                                <div className={"flex items-center justify-center py-10"}>
                                    <div className={"w-10 h-10 border-4 border-t-white border-stone-600 rounded-full animate-spin"}/>
                                </div>
                            ) : (
                                <div className={"flex flex-col"}>
                                    <div className={"flex flex-row items-center gap-5 justify-between"}>
                                        <div className={"flex flex-row items-center gap-2"}>
                                            <p className={"text-xl font-semibold"}>Proxy Server</p>
                                            <div className={`w-2 h-2 rounded-full animate-pulse ${proxyServerStatus.data.running ? "bg-green-500" : "bg-red-500"}`}/>
                                        </div>
                                        <div className={"flex items-center justify-center"}>
                                            <motion.a
                                                href={"/settings"}
                                                whileHover={{ scale: 1.05 }}
                                                whileTap={{ scale: 0.95 }}
                                                className={"p-2 rounded-full bg-white text-black"}
                                            >
                                                <ArrowUpRight size={15}/>
                                            </motion.a>
                                        </div>
                                    </div>
                                    {proxyServerStatus.data.running ? (
                                        <div className={"flex flex-row gap-2"}>
                                            <p className={"text-md"}>Running on port</p><p className={"text-amber-500"}>8080</p>
                                        </div>
                                    ) : (
                                        <p className={"text-md"}>Not running</p>
                                    )}
                                    <div className={"flex flex-row items-center gap-2 mt-5 justify-between"}>
                                        <motion.button
                                            whileHover={{ scale: 1.05 }}
                                            whileTap={{ scale: 0.95 }}
                                            className={"px-3 py-1 bg-white hover:bg-white/80 transition-colors duration-200 rounded-4xl text-sm font-medium text-black"}
                                        >
                                            Restart
                                        </motion.button>
                                        {proxyServerStatus.data.running ? (
                                            <motion.button
                                                whileHover={{ scale: 1.05 }}
                                                whileTap={{ scale: 0.95 }}
                                                className={"px-3 py-1 bg-red-500 hover:bg-red-500/80 transition-colors duration-200 rounded-4xl text-sm font-medium text-white"}
                                            >
                                                Stop
                                            </motion.button>
                                        ) : (
                                            <motion.button
                                                whileHover={{ scale: 1.05 }}
                                                whileTap={{ scale: 0.95 }}
                                                className={"px-3 py-1 bg-green-500 hover:bg-green-500/80 transition-colors duration-200 rounded-4xl text-sm font-medium text-white"}
                                            >
                                                Start
                                            </motion.button>
                                        )}
                                    </div>
                                </div>
                            )}
                    </div>
                </div>
                <div className={"border-box lg:w-2/5"}>
                    <div className={"border-box bg-stone-800 rounded-4xl p-5"}>
                        <div className={"flex flex-col gap-5"}>
                            <div className={"flex flex-row items-center gap-5 justify-between"}>
                                <p className={"text-xl font-semibold"}>Plugins</p>
                                <div className={"flex items-center justify-center"}>
                                    <motion.a
                                        href={"/plugins"}
                                        whileHover={{ scale: 1.05 }}
                                        whileTap={{ scale: 0.95 }}
                                        className={"p-2 rounded-full bg-white text-black"}
                                    >
                                        <ArrowUpRight size={15}/>
                                    </motion.a>
                                </div>
                            </div>
                            {pluginPaginatedData.loading ? (
                                <div className={"flex items-center justify-center py-10"}>
                                    <div className={"w-10 h-10 border-4 border-t-white border-stone-600 rounded-full animate-spin"}/>
                                </div>
                            ) : (
                                <div>
                                    {pluginPaginatedData.data.length === 0 ? (
                                        <div className={"flex flex-col items-center justify-center py-10"}>
                                            <p className={"text-lg"}>You didn't create any plugins yet</p>
                                            <a href={"/plugins/new"} className={"text-lg underline"}>Start with creating one</a>
                                        </div>
                                    ) : (
                                        <div className={"grid grid-cols-1 gap-2"}>
                                            {pluginPaginatedData.data.map((plugin, idx) => (
                                                <motion.div
                                                    key={plugin.id}
                                                    initial={{ opacity: 0, y: 10 }}
                                                    animate={{ opacity: 1, y: 0 }}
                                                    transition={{ duration: 0.2, delay: idx * 0.1 }}
                                                    className={"col-span-1 flex flex-row items-center justify-between p-1 bg-amber-500 rounded-2xl"}
                                                >
                                                    <div className={"flex flex-col items-start justify-center px-4"}>
                                                        <p className={"text-sm"}>Name</p>
                                                        <p className={"text-md font-semibold"}>{plugin.name}</p>
                                                    </div>
                                                    <div className={"flex items-center justify-center h-ful"}>
                                                        <motion.a
                                                            whileHover={{ scale: 1.05 }}
                                                            whileTap={{ scale: 0.95 }}
                                                            href={`/plugins/plugin?name=${plugin.name}`}
                                                            className={"px-3 py-3 rounded-2xl bg-white text-black"}
                                                        >
                                                            <ChevronRight size={15} />
                                                        </motion.a>
                                                    </div>
                                                </motion.div>
                                            ))}
                                        </div>
                                    )}
                                </div>
                            )}
                        </div>
                    </div>
                </div>
                <div className={"border-box lg:w-2/5"}>
                    <div className={"border-box bg-stone-800 rounded-4xl p-5"}>
                        <div className={"flex flex-col gap-5"}>
                            <div className={"flex flex-row items-center gap-5 justify-between"}>
                                <p className={"text-xl font-semibold"}>Active Routes</p>
                                <div className={"flex items-center justify-center"}>
                                    <motion.a
                                        href={"/routes"}
                                        whileHover={{ scale: 1.05 }}
                                        whileTap={{ scale: 0.95 }}
                                        className={"p-2 rounded-full bg-white text-black"}
                                    >
                                        <ArrowUpRight size={15}/>
                                    </motion.a>
                                </div>
                            </div>
                            {routePaginatedData.loading ? (
                                <div className={"flex items-center justify-center py-10"}>
                                    <div className={"w-10 h-10 border-4 border-t-white border-stone-600 rounded-full animate-spin"}/>
                                </div>
                            ) : (
                                <div>
                                    {routePaginatedData.data.length === 0 ? (
                                        <div className={"flex flex-col items-center justify-center py-10"}>
                                            <p className={"text-lg"}>No active routes found</p>
                                            <a href={"/routes"} className={"text-lg underline"}>Manage routes</a>
                                        </div>
                                    ) : (
                                        <div className={"grid grid-cols-1 gap-2"}>
                                            {routePaginatedData.data.map((route, idx) => (
                                                <motion.div
                                                    key={route.id}
                                                    initial={{ opacity: 0, y: 10 }}
                                                    animate={{ opacity: 1, y: 0 }}
                                                    transition={{ duration: 0.2, delay: idx * 0.1 }}
                                                    className={"col-span-1 flex flex-row items-center justify-between p-1 bg-amber-500 rounded-2xl"}
                                                >
                                                    <div className={"flex flex-col items-start justify-center px-4"}>
                                                        <p className={"text-sm"}>Path</p>
                                                        <p className={"text-md font-semibold"}>{route.path}</p>
                                                    </div>
                                                    <div className={"flex items-center justify-center h-ful"}>
                                                        <motion.a
                                                            whileHover={{ scale: 1.05 }}
                                                            whileTap={{ scale: 0.95 }}
                                                            href={`/routes/routes?path=${route.path}`}
                                                            className={"px-3 py-3 rounded-2xl bg-white text-black"}
                                                        >
                                                            <ChevronRight size={15} />
                                                        </motion.a>
                                                    </div>
                                                </motion.div>
                                            ))}
                                        </div>
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