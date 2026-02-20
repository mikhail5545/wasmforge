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
import React, {useState, useCallback, useEffect} from "react";
import {useData} from "@/hooks/useData";
import {
    BookKey,
    CalendarClock,
    Lock,
    File,
    Trash,
    ChevronRight, Pencil, Plus,
} from "lucide-react";
import {motion} from "motion/react";
import {Scrollbar} from "react-scrollbars-custom";
import {usePaginatedData} from "@/hooks/usePaginatedData";
import {ModalDialog} from "@/components/dialog/ModalDialog";
import PageLayout from "@/components/layout/PageLayout";
import {useMutation} from "@/hooks/useMutation";


export default function PluginPage() {
    const links = [
        { label: "Routes", href: "/routes", active: false },
        { label: "Plugins", href: "/plugins", active: false },
        { label: "Settings", href: "/settings", active: false },
    ];

    useEffect(() => {
        document.title = "Plugin Details - WasmForge";
    });

    const params = useSearchParams();
    const name = params.get("name") || "";
    const pluginDetails = useData<WasmForge.Plugin>(`http://localhost:8080/api/plugins/${name}`, "plugin");

    const routePluginsPaginatedData = usePaginatedData<WasmForge.RoutePlugin>(
        `api/route-plugins?p_ids=${pluginDetails.data?.id}`,
        "route_plugins",
        5,
        "created_at",
        "desc",
        { preload: true },
    );

    const routesPaginatedData = usePaginatedData<WasmForge.Route>(
        `api/routes`,
        "routes",
        10,
        "created_at",
        "desc",
        { preload: false },
    );
    const router = useRouter();
    const [showNewAssociationModal, setShowNewAssociationModal] = useState(false);
    const [selectedRouteId, setSelectedRouteId] = useState<string | null>(null);
    const [showDeleteConfirmation, setShowDeleteConfirmation] = useState(false);

    const mutation = useMutation();

    const deletePlugin = useCallback(
        async () => {
            if(!pluginDetails.data) return;

            const response = await mutation.mutate(`http://localhost:8080/api/plugins/${pluginDetails.data.id}`, "DELETE");
            if(response.success) {
                router.push("/plugins");
            } else {
                alert("Failed to delete plugin");
            }
        }, [pluginDetails.data, mutation, router]
    );

    const globalError = pluginDetails.error || routePluginsPaginatedData.error || routesPaginatedData.error;
    const unsetError = async () => {
        mutation.setError(null);
        if (pluginDetails.error) {
            await pluginDetails.refetch();
        }
        if (routePluginsPaginatedData.error) {
            await routePluginsPaginatedData.refetch();
        }
        if (routesPaginatedData.error) {
            await routesPaginatedData.refetch();
        }
    };

    return (
        <PageLayout>
            <NavBar links={links} />
            <div className={"flex flex-col gap-5 w-full mt-20"}>
                <div className={"flex flex-row w-full lg:w-2/3 gap-5"}>
                    <div className={"flex flex-row px-4 items-center justify-between bg-stone-800 rounded-4xl p-3 w-2/3"}>
                        <p className={"text-xl font-semibold"}>Plugin Details</p>
                        <div className={"flex flex-row gap-2"}>
                            <motion.button
                                whileHover={{ scale: 1.05 }}
                                whileTap={{ scale: 0.95 }}
                                className={"px-2 py-2 rounded-full bg-white text-black hover:bg-white/80 transition-colors duration-200"}
                            >
                                <Pencil size={15}/>
                            </motion.button>
                            <motion.button
                                whileHover={{ scale: 1.05 }}
                                whileTap={{ scale: 0.95 }}
                                onClick={() => setShowDeleteConfirmation(true)}
                                className={"px-2 py-2 rounded-full bg-red-500 text-white hover:bg-red-500/80 transition-colors duration-200"}
                            >
                                <Trash size={15}/>
                            </motion.button>
                        </div>
                    </div>
                </div>
                <ModalDialog title={"Error"} visible={!!globalError} onClose={unsetError} >
                    <div className={"flex flex-col gap-5"}>
                        <p className={"text-md font-semibold"}>{globalError?.message}</p>
                        <Scrollbar style={{ height: 200 }}>
                            <p className={"text-md"}>{globalError?.details}</p>
                        </Scrollbar>
                        <div className={"flex flex-row items-end justify-end"}>
                            <motion.button
                                whileHover={{ scale: 1.05 }}
                                whileTap={{ scale: 0.95 }}
                                onClick={unsetError}
                                className={"px-3 py-2 rounded-full bg-white text-black hover:bg-white/80 transition-colors duration-200"}
                            >
                                Close
                            </motion.button>
                        </div>
                    </div>
                </ModalDialog>
                <ModalDialog title={"Delete Plugin"} visible={showDeleteConfirmation} onClose={() => setShowDeleteConfirmation(false)} >
                    <div className={"flex flex-col gap-5"}>
                        <p className={"text-md"}>Are you sure? This action is irreversible.</p>
                        <div className={"flex flex-row items-end justify-end"}>
                            <motion.button
                                whileHover={{ scale: 1.05 }}
                                whileTap={{ scale: 0.95 }}
                                onClick={deletePlugin}
                                className={"px-3 py-2 rounded-full bg-red-500 text-white hover:bg-red-500/80 transition-colors duration-200"}
                            >
                                Delete Plugin
                            </motion.button>
                        </div>
                    </div>
                </ModalDialog>
                <ModalDialog
                    visible={showNewAssociationModal}
                    onClose={() => setShowNewAssociationModal(false)}
                    title={"Select Route"}
                >
                    <div className={"flex flex-col gap-5 mt-5"}>
                        {routesPaginatedData.loading ? (
                            <div className={"flex items-center justify-center py-10"}>
                                <div className={"w-10 h-10 border-4 border-t-white border-stone-800 rounded-full animate-spin"}/>
                            </div>
                        ) : (
                            <div>
                                {routesPaginatedData.data.length === 0 ? (
                                    <div className={"flex flex-col items-center justify-center py-10"}>
                                        <p className={"text-lg"}>You didn't create any routes yet</p>
                                        <a href={"/routes/new"} className={"text-lg underline"}>Start with creating one</a>
                                    </div>
                                ) : (
                                    <Scrollbar style={{ height: 500 }}>
                                        <div className={"grid grid-cols-1 gap-2 pr-2"}>
                                            {routesPaginatedData.data.map((route, idx) => (
                                                <motion.div
                                                    key={route.id}
                                                    initial={{ opacity: 0, y: 10 }}
                                                    animate={{ opacity: 1, y: 0 }}
                                                    transition={{ duration: 0.2, delay: idx * 0.1 }}
                                                    onClick={() => setSelectedRouteId(route.id)}
                                                    className={`px-4 col-span-1 flex flex-row items-center p-1 rounded-2xl ${
                                                        selectedRouteId === route.id ? "bg-amber-500" : "border border-amber-500 hover:bg-amber-500 transition-colors duration-200"
                                                    }`}
                                                >
                                                    <div className={"flex flex-col items-start justify-center w-1/3"}>
                                                        <p className={"text-sm"}>Path</p>
                                                        <p className={"text-md font-semibold truncate"}>{route.path}</p>
                                                    </div>
                                                    <div className={"flex flex-col items-start justify-center w-2/3"}>
                                                        <p className={"text-sm"}>Target URL</p>
                                                        <p className={"text-md font-semibold truncate"}>{route.target_url}</p>
                                                    </div>
                                                </motion.div>
                                            ))}
                                        </div>
                                    </Scrollbar>
                                )}
                                {routesPaginatedData.nextPageToken && (
                                    <div className={"flex items-center justify-center pt-3"}>
                                        <motion.button
                                            whileHover={{ scale: 1.05 }}
                                            whileTap={{ scale: 0.95 }}
                                            onClick={() => routesPaginatedData.nextPage(routesPaginatedData.nextPageToken, {append: true, force: true})}
                                            className={"text-sm underline"}
                                        >
                                            Load more
                                        </motion.button>
                                    </div>
                                )}
                                <div className={"flex flex-row items-center justify-end pt-5"}>
                                    <div className={"flex flex-row items-center justify-between gap-5"}>
                                        <motion.button
                                            disabled={routesPaginatedData.loading}
                                            whileHover={{ scale: 1.05 }}
                                            whileTap={{ scale: 0.95 }}
                                            onClick={() => {router.push(`/routes/plugins/new?plugin_id=${pluginDetails.data.id}`)}}
                                            className={"px-3 py-2 rounded-full bg-white text-black hover:bg-white/80 transition-colors duration-200 disabled:opacity-70 disabled:cursor-not-allowed"}
                                        >
                                            Choose later
                                        </motion.button>
                                        <motion.button
                                            disabled={!selectedRouteId}
                                            whileHover={{ scale: 1.05 }}
                                            whileTap={{ scale: 0.95 }}
                                            onClick={() => {router.push(`/routes/plugins/new?plugin_id=${pluginDetails.data.id}&route_id=${selectedRouteId}`)}}
                                            className={"px-3 py-2 rounded-full bg-white text-black hover:bg-white/80 transition-colors duration-200 disabled:opacity-70 disabled:cursor-not-allowed"}
                                        >
                                            Submit
                                        </motion.button>
                                    </div>
                                </div>
                            </div>
                        )}
                    </div>
                </ModalDialog>
                <div className={"flex flex-col lg:flex-row gap-5 w-full"}>
                    <div className={"w-full lg:w-2/3"}>
                        <div className={"border-box w-full bg-stone-800 rounded-4xl p-5"}>
                            {pluginDetails.loading ? (
                                <div className={"flex items-center justify-center py-20"}>
                                    <div className={"w-10 h-10 border-4 border-t-white border-stone-800 rounded-full animate-spin"}/>
                                </div>
                            ) : (
                                <div className={"flex flex-col gap-5"}>
                                    <div className={"flex flex-row w-full gap-5"}>
                                        <div className={"flex items-center justify-center"}>
                                            <div className={"bg-amber-500 p-2 rounded-full"}>
                                                <BookKey size={15}/>
                                            </div>
                                        </div>
                                        <div className={"flex flex-col"}>
                                            <p className={"text-md"}>Name</p>
                                            <p className={"text-md font-semibold"}>{pluginDetails.data.name}</p>
                                        </div>
                                    </div>
                                    <div className={"flex flex-row w-full gap-5"}>
                                        <div className={"flex items-center justify-center"}>
                                            <div className={"bg-amber-500 p-2 rounded-full"}>
                                                <File size={15}/>
                                            </div>
                                        </div>
                                        <div className={"flex flex-col"}>
                                            <p className={"text-md"}>Filename</p>
                                            <p className={"text-md font-semibold"}>{pluginDetails.data.filename}</p>
                                        </div>
                                    </div>
                                    <div className={"flex flex-row w-full gap-5"}>
                                        <div className={"flex items-center justify-center"}>
                                            <div className={"bg-amber-500 p-2 rounded-full"}>
                                                <Lock size={15}/>
                                            </div>
                                        </div>
                                        <div className={"flex flex-col"}>
                                            <p className={"text-md"}>Checksum</p>
                                            <p className={"text-md font-semibold"}>{`SHA-256:${pluginDetails.data.checksum}`}</p>
                                        </div>
                                    </div>
                                    <div className={"flex flex-row w-full gap-5"}>
                                        <div className={"flex items-center justify-center"}>
                                            <div className={"bg-amber-500 p-2 rounded-full"}>
                                                <CalendarClock size={15}/>
                                            </div>
                                        </div>
                                        <div className={"flex flex-col"}>
                                            <p className={"text-md"}>Created at</p>
                                            <p className={"text-md font-semibold"}>{new Date(pluginDetails.data.created_at).toLocaleString()}</p>
                                        </div>
                                    </div>
                                </div>
                            )}
                        </div>
                    </div>
                    <div className={"w-full lg:w-1/3"}>
                        <div className={"border-box bg-stone-800 rounded-4xl p-5"}>
                            <div className={"flex flex-col gap-5"}>
                                <div className={"flex flex-row items-center gap-5 justify-between"}>
                                    <p className={"text-xl font-semibold"}>Associations</p>
                                    <div className={"flex items-center justify-center"}>
                                        <motion.button
                                            whileHover={{ scale: 1.05 }}
                                            whileTap={{ scale: 0.95 }}
                                            onClick={async () => {await routesPaginatedData.refetch(); setShowNewAssociationModal(true)}}
                                            className={"p-2 rounded-full bg-white text-black"}
                                        >
                                            <Plus size={15}/>
                                        </motion.button>
                                    </div>
                                </div>
                                {routePluginsPaginatedData.loading ? (
                                    <div className={"flex items-center justify-center py-10"}>
                                        <div className={"w-10 h-10 border-4 border-t-white border-stone-800 rounded-full animate-spin"}/>
                                    </div>
                                ) : (
                                    <div>
                                        {routePluginsPaginatedData.data.length === 0 ? (
                                            <div className={"flex flex-col items-center justify-center py-10"}>
                                                <p className={"text-lg"}>You didn't create any plugins yet</p>
                                                <motion.button
                                                    whileHover={{ scale: 1.05 }}
                                                    whileTap={{ scale: 0.95 }}
                                                    className={"text-lg underline"}
                                                >Start with creating one</motion.button>
                                            </div>
                                        ) : (
                                            <div className={"grid grid-cols-1 gap-2"}>
                                                {routePluginsPaginatedData.data.map((plugin, idx) => (
                                                    <motion.div
                                                        key={plugin.id}
                                                        initial={{ opacity: 0, y: 10 }}
                                                        animate={{ opacity: 1, y: 0 }}
                                                        transition={{ duration: 0.2, delay: idx * 0.1 }}
                                                        className={"col-span-1 flex flex-row items-center justify-between p-1 bg-amber-500 rounded-2xl"}
                                                    >
                                                        <div className={"flex flex-col items-start justify-center px-4"}>
                                                            <p className={"text-sm"}>Route ID</p>
                                                            <p className={"text-md font-semibold"}>{plugin.route_id}</p>
                                                        </div>
                                                        <div className={"flex items-center justify-center h-ful"}>
                                                            <motion.a
                                                                whileHover={{ scale: 1.05 }}
                                                                whileTap={{ scale: 0.95 }}
                                                                href={`/routes/plugins/plugin?id=${plugin.id}`}
                                                                className={"px-3 py-3 rounded-2xl bg-white text-black"}
                                                            >
                                                                <ChevronRight size={15} />
                                                            </motion.a>
                                                        </div>
                                                    </motion.div>
                                                ))}
                                                {routePluginsPaginatedData.nextPageToken && (
                                                    <div className={"flex items-center justify-center pt-5"}>
                                                        <motion.button
                                                            whileHover={{ scale: 1.05 }}
                                                            whileTap={{ scale: 0.95 }}
                                                            onClick={() => routePluginsPaginatedData.nextPage(routePluginsPaginatedData.nextPageToken, {append: true})}
                                                            className={"text-sm underline"}
                                                        >
                                                            Load more
                                                        </motion.button>
                                                    </div>
                                                )}
                                            </div>
                                        )}
                                    </div>
                                )}
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </PageLayout>
    );
}