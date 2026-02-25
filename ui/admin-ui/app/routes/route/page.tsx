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
import {
    CalendarClock,
    Link2,
    Route,
    Plus,
    Trash,
    Power,
    Handshake,
    Hourglass,
    Antenna,
    X, ChevronRight, Pencil, Check,
} from "lucide-react";
import React, {useEffect, useState, useCallback} from "react";
import {AnimatePresence,motion} from "motion/react";
import {usePaginatedData} from "@/hooks/usePaginatedData";
import PageLayout from "@/components/layout/PageLayout";
import Scrollbar from "react-scrollbars-custom";
import {ModalDialog} from "@/components/dialog/ModalDialog";
import {useMutation} from "@/hooks/useMutation";
import {Input} from "@headlessui/react";


function RoutePageContent() {
    const params = useSearchParams();
    const path = !params.get("path") ? "" : params.get("path")!;
    const routeDetails = useData<WasmForge.Route>(`http://localhost:8080/api/routes/${encodeURIComponent(path)}`, "route");
    const routePluginsPaginatedData = usePaginatedData<WasmForge.RoutePlugin>(
        `/api/route-plugins?r_ids=${routeDetails?.data?.id}`,
        "route_plugins",
        5,
        "created_at",
        "desc",
        { preload: true }
    );
    const pluginPaginatedData = usePaginatedData<WasmForge.Plugin>(
        "/api/plugins",
        "plugins",
        20,
        "created_at",
        "desc",
        { preload: false }
    );

    useEffect(() => {
        document.title = "Route Details - WasmForge";
    });

    const [showNewPluginDialog, setShowNewPluginDialog] = useState(false);
    const [selectedPluginId, setSelectedPluginId] = useState<string | null>(null);
    const [showDeleteConfirmation, setShowDeleteConfirmation] = useState(false);
    const [editMode, setEditMode] = useState(false);
    const [editableRoute, setEditableRoute] = useState<Omit<WasmForge.Route, "id" | "created_at"> | null>(null);
    const router = useRouter();

    const openNewPluginDialog = async () => {
        await pluginPaginatedData.refetch();
        setShowNewPluginDialog(true);
    };

    const mutation = useMutation();

    const deleteRoute = useCallback(
        async () => {
            if (!routeDetails.data) return;

            const response = await mutation.mutate(`http://localhost:8080/api/routes/${routeDetails.data.id}`, "DELETE");
            if (response.success) {
                router.push("/routes");
            } else {
            }
        }, [routeDetails.data, mutation, router]
    );

    const enableRoute = useCallback(
        async () => {
            if (!routeDetails.data) return;

            const response = await mutation.mutate(`http://localhost:8080/api/routes/${routeDetails.data.id}/enable`, "POST");
            if (response.success) {
                await routeDetails.refetch();
            } else {
            }
        }, [routeDetails, mutation]
    );

    const disableRoute = useCallback(
        async () => {
            if (!routeDetails.data) return;

            const response = await mutation.mutate(`http://localhost:8080/api/routes/${routeDetails.data.id}/disable`, "POST");
            if (response.success) {
                await routeDetails.refetch();
            } else {
            }
        }, [routeDetails, mutation]
    );

    const handleSubmitChanges = useCallback(
        async () => {
            if (!routeDetails.data || !editableRoute) return;

            const res = await mutation.mutate(`http://localhost:8080/api/routes/${routeDetails.data.id}`, "PATCH", JSON.stringify(editableRoute));
            if (res.success) {
                setEditMode(false);
                setEditableRoute(null);
                await routeDetails.refetch();
            }
        }, [routeDetails, editableRoute, mutation]
    );

    const globalError = mutation.error || routeDetails.error || routePluginsPaginatedData.error || pluginPaginatedData.error;
    const unsetError = async () => {
        mutation.setError(null);
        if (routeDetails.error) {
            await routeDetails.refetch();
        }
        if (routePluginsPaginatedData.error) {
            await routePluginsPaginatedData.refetch();
        }
        if (pluginPaginatedData.error) {
            await pluginPaginatedData.refetch();
        }
    };

    return (
        <div className={"flex flex-col gap-5 w-full mt-20"}>
            <div className={"flex flex-row w-full lg:w-2/3 gap-5"}>
                <div className={"flex flex-row px-4 items-center justify-between bg-stone-800 rounded-4xl p-3 w-2/3"}>
                    <p className={"text-xl font-semibold"}>Route Details</p>
                    <div className={"flex flex-row gap-2"}>
                        <AnimatePresence mode={"wait"}>
                            {editMode && (
                                <motion.button
                                    whileHover={{ scale: 1.05 }}
                                    whileTap={{ scale: 0.95 }}
                                    onClick={handleSubmitChanges}
                                    disabled={routeDetails.loading || mutation.loading}
                                    className={"p-2 rounded-full  text-white hover:bg-white/5 transition-colors duration-200 disabled:opacity-70 disabled:cursor-not-allowed"}
                                >
                                    {mutation.loading ? (
                                        <div className={"flex items-center justify-center"}>
                                            <div className={"w-4 h-4 border-2 border-t-black border-white rounded-full animate-spin"}/>
                                        </div>
                                    ) : (
                                        <Check size={15}/>
                                    )}
                                </motion.button>
                            )}
                        </AnimatePresence>
                        <motion.button
                            whileHover={{ scale: 1.05 }}
                            whileTap={{ scale: 0.95 }}
                            onClick={editMode ? () => setEditMode(false) : () => {setEditableRoute(routeDetails.data || {}); setEditMode(true)}}
                            disabled={routeDetails.loading || mutation.loading}
                            className={"p-2 rounded-full  text-white hover:bg-white/5 transition-colors duration-200 disabled:opacity-70 disabled:cursor-not-allowed"}
                        >
                            {editMode ? <X size={15}/> : <Pencil size={15}/>}
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
                <div className={"flex flex-row px-4 bg-stone-800 rounded-4xl p-3 w-1/3 lg:mr-3"}>
                    {routeDetails.loading ? (
                        <div className={"flex items-center justify-center"}>
                            <div className={"w-5 h-5 border-4 border-t-white border-stone-800 rounded-full animate-spin"}/>
                        </div>
                    ) : (
                        <div className={"flex flex-row items-center justify-between gap-2 w-full"}>
                            <p className={"text-xl"}>{routeDetails.data.enabled ? "Enabled" : "Disabled"}</p>
                            <motion.button
                                whileHover={{ scale: 1.05 }}
                                whileTap={{ scale: 0.95 }}
                                onClick={routeDetails.data.enabled ? disableRoute : enableRoute}
                                className={`px-2 py-2 rounded-full ${routeDetails.data.enabled ? "bg-red-500 hover:bg-red-500/80" : "bg-green-500 hover:bg-green-500/80"} transition-colors duration-200`}>
                                <Power size={15}/>
                            </motion.button>
                        </div>
                    )}
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
            <ModalDialog title={"Delete Route"} visible={showDeleteConfirmation} onClose={() => setShowDeleteConfirmation(false)} >
                <div className={"flex flex-col gap-5"}>
                    <p className={"text-md"}>Are you sure? This action is irreversible.</p>
                    <div className={"flex flex-row items-end justify-end"}>
                        <motion.button
                            whileHover={{ scale: 1.05 }}
                            whileTap={{ scale: 0.95 }}
                            onClick={deleteRoute}
                            className={"px-3 py-2 rounded-full bg-red-500 text-white hover:bg-red-500/80 transition-colors duration-200"}
                        >
                            Delete Route
                        </motion.button>
                    </div>
                </div>
            </ModalDialog>
            <AnimatePresence mode={"popLayout"}>
                {showNewPluginDialog && (
                    <motion.div
                        initial={{ opacity: 0 }}
                        animate={{ opacity: 1 }}
                        exit={{ opacity: 0 }}
                        className={"backdrop-blur-2xl fixed inset-0 flex items-start justify-center z-50"}
                    >
                        <motion.div
                            key={"new-plugin-dialog"}
                            initial={{ opacity: 0, y: -20 }}
                            animate={{ opacity: 1, y: 0 }}
                            exit={{ opacity: 0, y: -20 }}
                            transition={{ duration: 0.3 }}
                            className={"fixed top-20 left-1/2 transform -translate-x-1/2 w-full md:w-1/2 lg:w-1/3 bg-stone-800 rounded-4xl z-50 border-box p-5"}
                        >
                            <div className={"flex flex-col gap-5"}>
                                <div className={"flex flex-row items-center gap-5 justify-between"}>
                                    <p className={"text-xl font-semibold"}>Select plugin</p>
                                    <div className={"flex items-center justify-center"}>
                                        <motion.button
                                            whileHover={{ scale: 1.05 }}
                                            whileTap={{ scale: 0.95 }}
                                            onClick={() => {setShowNewPluginDialog(false); setSelectedPluginId(null)}}
                                            className={"p-2 rounded-full bg-white text-black"}
                                        >
                                            <X size={15}/>
                                        </motion.button>
                                    </div>
                                </div>
                                {pluginPaginatedData.loading ? (
                                    <div className={"flex items-center justify-center py-10"}>
                                        <div className={"w-10 h-10 border-4 border-t-white border-stone-800 rounded-full animate-spin"}/>
                                    </div>
                                ) : (
                                    <div>
                                        {pluginPaginatedData.data.length === 0 ? (
                                            <div className={"flex flex-col items-center justify-center py-10"}>
                                                <p className={"text-lg"}>You didn't create any plugins yet</p>
                                                <a href={"/plugins/new"} className={"text-lg underline"}>Start with creating one</a>
                                            </div>
                                        ) : (
                                            <Scrollbar style={{ height: 500 }}>
                                                <div className={"grid grid-cols-1 gap-2 pr-2"}>
                                                    {pluginPaginatedData.data.map((plugin, idx) => (
                                                        <motion.div
                                                            key={plugin.id}
                                                            initial={{ opacity: 0, y: 10 }}
                                                            animate={{ opacity: 1, y: 0 }}
                                                            transition={{ duration: 0.2, delay: idx * 0.1 }}
                                                            onClick={() => setSelectedPluginId(plugin.id)}
                                                            className={`px-4 col-span-1 flex flex-row items-center p-1 rounded-2xl ${
                                                                selectedPluginId === plugin.id ? "bg-amber-500" : "border border-amber-500 hover:bg-amber-500 transition-colors duration-200"
                                                            }`}
                                                        >
                                                            <div className={"flex flex-col items-start justify-center w-1/2"}>
                                                                <p className={"text-sm"}>Name</p>
                                                                <p className={"text-md font-semibold truncate"}>{plugin.name}</p>
                                                            </div>
                                                            <div className={"flex flex-col items-start justify-center w-1/2"}>
                                                                <p className={"text-sm"}>Filename</p>
                                                                <p className={"text-md font-semibold truncate"}>{plugin.filename}</p>
                                                            </div>
                                                        </motion.div>
                                                    ))}
                                                </div>
                                            </Scrollbar>
                                        )}
                                        {pluginPaginatedData.nextPageToken && (
                                            <div className={"flex items-center justify-center pt-3"}>
                                                <motion.button
                                                    whileHover={{ scale: 1.05 }}
                                                    whileTap={{ scale: 0.95 }}
                                                    onClick={() => pluginPaginatedData.nextPage(pluginPaginatedData.nextPageToken, {append: true, force: true})}
                                                    className={"text-sm underline"}
                                                >
                                                    Load more
                                                </motion.button>
                                            </div>
                                        )}
                                        <div className={"flex flex-row items-center justify-end pt-5"}>
                                            <div className={"flex flex-row items-center justify-between gap-5"}>
                                                <motion.button
                                                    disabled={pluginPaginatedData.loading}
                                                    whileHover={{ scale: 1.05 }}
                                                    whileTap={{ scale: 0.95 }}
                                                    onClick={() => {router.push(`/routes/plugins/new?route_id=${routeDetails.data.id}`)}}
                                                    className={"px-3 py-2 rounded-full bg-white text-black hover:bg-white/80 transition-colors duration-200 disabled:opacity-70 disabled:cursor-not-allowed"}
                                                >
                                                    Choose later
                                                </motion.button>
                                                <motion.button
                                                    disabled={!selectedPluginId}
                                                    whileHover={{ scale: 1.05 }}
                                                    whileTap={{ scale: 0.95 }}
                                                    onClick={() => {router.push(`/routes/plugins/new?route_id=${routeDetails.data.id}&plugin_id=${selectedPluginId}`)}}
                                                    className={"px-3 py-2 rounded-full bg-white text-black hover:bg-white/80 transition-colors duration-200 disabled:opacity-70 disabled:cursor-not-allowed"}
                                                >
                                                    Submit
                                                </motion.button>
                                            </div>
                                        </div>
                                    </div>
                                )}
                            </div>
                        </motion.div>
                    </motion.div>
                )}
            </AnimatePresence>
            <div className={"flex flex-col lg:flex-row gap-5 w-full"}>
                <div className={"w-full lg:w-1/3"}>
                    <div className={"border-box w-full bg-stone-800 rounded-4xl p-5"}>
                        {routeDetails.loading ? (
                            <div className={"flex items-center justify-center py-20"}>
                                <div className={"w-10 h-10 border-4 border-t-white border-stone-800 rounded-full animate-spin"}/>
                            </div>
                        ) : (
                            <div className={"flex flex-col gap-5"}>
                                <div className={"flex flex-row w-full gap-5"}>
                                    <div className={"flex items-center justify-center"}>
                                        <div className={"bg-amber-500 p-2 rounded-full"}>
                                            <Route size={15}/>
                                        </div>
                                    </div>
                                    <div className={"flex flex-col gap-1 w-full"}>
                                        <p className={"text-md px-2"}>Path</p>
                                        <AnimatePresence mode={"wait"}>
                                            {editMode ? (
                                                <motion.div
                                                    key={"edit"}
                                                    initial={{ opacity: 0, y: -10 }}
                                                    animate={{ opacity: 1, y: 0 }}
                                                    exit={{ opacity: 0, y: -10 }}
                                                    transition={{ duration: 0.2 }}
                                                >
                                                    <Input
                                                        required
                                                        type={"text"}
                                                        value={editableRoute?.path || ""}
                                                        onChange={(e) => setEditableRoute(prev => prev ? {...prev, path: e.target.value} : null)}
                                                        className={`w-full px-3 py-2 rounded-lg bg-stone-700 text-white focus:outline-none focus:ring-2 focus:ring-amber-500`}
                                                    />
                                                </motion.div>
                                            ) : (
                                                <motion.p
                                                    key={"view"}
                                                    initial={{ opacity: 0, y: -10 }}
                                                    animate={{ opacity: 1, y: 0 }}
                                                    exit={{ opacity: 0, y: -10 }}
                                                    transition={{ duration: 0.2 }}
                                                    className={"text-md font-semibold px-2"}
                                                >
                                                    {routeDetails.data.path}
                                                </motion.p>
                                            )}
                                        </AnimatePresence>
                                    </div>
                                </div>
                                <div className={"flex flex-row w-full gap-5"}>
                                    <div className={"flex items-center justify-center"}>
                                        <div className={"bg-amber-500 p-2 rounded-full"}>
                                            <Link2 size={15}/>
                                        </div>
                                    </div>
                                    <div className={"flex flex-col gap-1 w-full"}>
                                        <p className={"text-md px-2"}>Target URL</p>
                                        <AnimatePresence mode={"wait"}>
                                            {editMode ? (
                                                <motion.div
                                                    key={"edit"}
                                                    initial={{ opacity: 0, y: -10 }}
                                                    animate={{ opacity: 1, y: 0 }}
                                                    exit={{ opacity: 0, y: -10 }}
                                                    transition={{ duration: 0.2 }}
                                                >
                                                    <Input
                                                        required
                                                        type={"text"}
                                                        value={editableRoute?.target_url || ""}
                                                        onChange={(e) => setEditableRoute(prev => prev ? {...prev, target_url: e.target.value} : null)}
                                                        className={`w-full px-3 py-2 rounded-lg bg-stone-700 text-white focus:outline-none focus:ring-2 focus:ring-amber-500`}
                                                    />
                                                </motion.div>
                                            ) : (
                                                <motion.p
                                                    key={"view"}
                                                    initial={{ opacity: 0, y: -10 }}
                                                    animate={{ opacity: 1, y: 0 }}
                                                    exit={{ opacity: 0, y: -10 }}
                                                    transition={{ duration: 0.2 }}
                                                    className={"text-md font-semibold px-2"}
                                                >
                                                    {routeDetails.data.target_url}
                                                </motion.p>
                                            )}
                                        </AnimatePresence>
                                    </div>
                                </div>
                                <div className={"flex flex-row w-full gap-5"}>
                                    <div className={"flex items-center justify-center"}>
                                        <div className={"bg-amber-500 p-2 rounded-full"}>
                                            <CalendarClock size={15}/>
                                        </div>
                                    </div>
                                    <div className={"flex flex-col gap-1"}>
                                        <p className={"text-md px-2"}>Created at</p>
                                        <p className={"text-md font-semibold px-2"}>{new Date(routeDetails.data.created_at).toLocaleString()}</p>
                                    </div>
                                </div>
                            </div>
                        )}
                    </div>
                </div>
                <div className={"w-full lg:w-1/3"}>
                    <div className={"border-box w-full bg-stone-800 rounded-4xl p-5"}>
                        {routeDetails.loading ? (
                            <div className={"flex items-center justify-center py-20"}>
                                <div className={"w-10 h-10 border-4 border-t-white border-stone-800 rounded-full animate-spin"}/>
                            </div>
                        ) : (
                            <div className={"flex flex-col gap-5"}>
                                <div className={"flex flex-row w-full gap-5"}>
                                    <div className={"flex items-center justify-center"}>
                                        <div className={"bg-amber-500 p-2 rounded-full"}>
                                            <Hourglass size={15}/>
                                        </div>
                                    </div>
                                    <div className={"flex flex-col gap-1 w-full"}>
                                        <p className={"text-md px-2"}>Idle connection timeout</p>
                                        <AnimatePresence mode={"wait"}>
                                            {editMode ? (
                                                <motion.div
                                                    key={"edit"}
                                                    initial={{ opacity: 0, y: -10 }}
                                                    animate={{ opacity: 1, y: 0 }}
                                                    exit={{ opacity: 0, y: -10 }}
                                                    transition={{ duration: 0.2 }}
                                                >
                                                    <Input
                                                        required
                                                        type={"number"}
                                                        value={editableRoute?.idle_conn_timeout || 0}
                                                        onChange={(e) => setEditableRoute(prev => prev ? {...prev, idle_conn_timeout: parseInt(e.target.value)} : null)}
                                                        className={`w-full px-3 py-2 rounded-lg bg-stone-700 text-white focus:outline-none focus:ring-2 focus:ring-amber-500`}
                                                    />
                                                </motion.div>
                                            ) : (
                                                <motion.p
                                                    key={"view"}
                                                    initial={{ opacity: 0, y: -10 }}
                                                    animate={{ opacity: 1, y: 0 }}
                                                    exit={{ opacity: 0, y: -10 }}
                                                    transition={{ duration: 0.2 }}
                                                    className={"text-md font-semibold px-2"}
                                                >
                                                    {routeDetails.data.idle_conn_timeout}
                                                </motion.p>
                                            )}
                                        </AnimatePresence>
                                    </div>
                                </div>
                                <div className={"flex flex-row w-full gap-5"}>
                                    <div className={"flex items-center justify-center"}>
                                        <div className={"bg-amber-500 p-2 rounded-full"}>
                                            <Handshake size={15}/>
                                        </div>
                                    </div>
                                    <div className={"flex flex-col gap-1 w-full"}>
                                        <p className={"text-md px-2"}>TLS handshake timeout</p>
                                        <AnimatePresence mode={"wait"}>
                                            {editMode ? (
                                                <motion.div
                                                    key={"edit"}
                                                    initial={{ opacity: 0, y: -10 }}
                                                    animate={{ opacity: 1, y: 0 }}
                                                    exit={{ opacity: 0, y: -10 }}
                                                    transition={{ duration: 0.2 }}
                                                >
                                                    <Input
                                                        required
                                                        type={"number"}
                                                        value={editableRoute?.tls_handshake_timeout || 0}
                                                        onChange={(e) => setEditableRoute(prev => prev ? {...prev, tls_handshake_timeout: parseInt(e.target.value)} : null)}
                                                        className={`w-full px-3 py-2 rounded-lg bg-stone-700 text-white focus:outline-none focus:ring-2 focus:ring-amber-500`}
                                                    />
                                                </motion.div>
                                            ) : (
                                                <motion.p
                                                    key={"view"}
                                                    initial={{ opacity: 0, y: -10 }}
                                                    animate={{ opacity: 1, y: 0 }}
                                                    exit={{ opacity: 0, y: -10 }}
                                                    transition={{ duration: 0.2 }}
                                                    className={"text-md font-semibold px-2"}
                                                >
                                                    {routeDetails.data.tls_handshake_timeout}
                                                </motion.p>
                                            )}
                                        </AnimatePresence>
                                    </div>
                                </div>
                                <div className={"flex flex-row w-full gap-5"}>
                                    <div className={"flex items-center justify-center"}>
                                        <div className={"bg-amber-500 p-2 rounded-full"}>
                                            <Hourglass size={15}/>
                                        </div>
                                    </div>
                                    <div className={"flex flex-col gap-1 w-full"}>
                                        <p className={"text-md px-2"}>Expect continue timeout</p>
                                        <AnimatePresence mode={"wait"}>
                                            {editMode ? (
                                                <motion.div
                                                    key={"edit"}
                                                    initial={{ opacity: 0, y: -10 }}
                                                    animate={{ opacity: 1, y: 0 }}
                                                    exit={{ opacity: 0, y: -10 }}
                                                    transition={{ duration: 0.2 }}
                                                >
                                                    <Input
                                                        required
                                                        type={"number"}
                                                        value={editableRoute?.expect_continue_timeout || 0}
                                                        onChange={(e) => setEditableRoute(prev => prev ? {...prev, expect_continue_timeout: parseInt(e.target.value)} : null)}
                                                        className={`w-full px-3 py-2 rounded-lg bg-stone-700 text-white focus:outline-none focus:ring-2 focus:ring-amber-500`}
                                                    />
                                                </motion.div>
                                            ) : (
                                                <motion.p
                                                    key={"view"}
                                                    initial={{ opacity: 0, y: -10 }}
                                                    animate={{ opacity: 1, y: 0 }}
                                                    exit={{ opacity: 0, y: -10 }}
                                                    transition={{ duration: 0.2 }}
                                                    className={"text-md font-semibold px-2"}
                                                >
                                                    {routeDetails.data.expect_continue_timeout}
                                                </motion.p>
                                            )}
                                        </AnimatePresence>
                                    </div>
                                </div>
                                <div className={"flex flex-row w-full gap-5"}>
                                    <div className={"flex items-center justify-center"}>
                                        <div className={"bg-amber-500 p-2 rounded-full"}>
                                            <Antenna size={15}/>
                                        </div>
                                    </div>
                                    <div className={"flex flex-col gap-1 w-full"}>
                                        <p className={"text-md px-2"}>Max connections per host</p>
                                        <AnimatePresence mode={"wait"}>
                                            {editMode ? (
                                                <motion.div
                                                    key={"edit"}
                                                    initial={{ opacity: 0, y: -10 }}
                                                    animate={{ opacity: 1, y: 0 }}
                                                    exit={{ opacity: 0, y: -10 }}
                                                    transition={{ duration: 0.2 }}
                                                >
                                                    <Input
                                                        required
                                                        type={"number"}
                                                        value={editableRoute?.max_cons_per_host || 0}
                                                        onChange={(e) => setEditableRoute(prev => prev ? {...prev, max_cons_per_host: parseInt(e.target.value)} : null)}
                                                        className={`w-full px-3 py-2 rounded-lg bg-stone-700 text-white focus:outline-none focus:ring-2 focus:ring-amber-500`}
                                                    />
                                                </motion.div>
                                            ) : (
                                                <motion.p
                                                    key={"view"}
                                                    initial={{ opacity: 0, y: -10 }}
                                                    animate={{ opacity: 1, y: 0 }}
                                                    exit={{ opacity: 0, y: -10 }}
                                                    transition={{ duration: 0.2 }}
                                                    className={"text-md font-semibold px-2"}
                                                >
                                                    {routeDetails.data.max_cons_per_host || "default"}
                                                </motion.p>
                                            )}
                                        </AnimatePresence>
                                    </div>
                                </div>
                                <div className={"flex flex-row w-full gap-5"}>
                                    <div className={"flex items-center justify-center"}>
                                        <div className={"bg-amber-500 p-2 rounded-full"}>
                                            <Antenna size={15}/>
                                        </div>
                                    </div>
                                    <div className={"flex flex-col gap-1 w-full"}>
                                        <p className={"text-md px-2"}>Max idle connections</p>
                                        <AnimatePresence mode={"wait"}>
                                            {editMode ? (
                                                <motion.div
                                                    key={"edit"}
                                                    initial={{ opacity: 0, y: -10 }}
                                                    animate={{ opacity: 1, y: 0 }}
                                                    exit={{ opacity: 0, y: -10 }}
                                                    transition={{ duration: 0.2 }}
                                                >
                                                    <Input
                                                        required
                                                        type={"number"}
                                                        value={editableRoute?.max_idle_cons || 0}
                                                        onChange={(e) => setEditableRoute(prev => prev ? {...prev, max_idle_cons: parseInt(e.target.value)} : null)}
                                                        className={`w-full px-3 py-2 rounded-lg bg-stone-700 text-white focus:outline-none focus:ring-2 focus:ring-amber-500`}
                                                    />
                                                </motion.div>
                                            ) : (
                                                <motion.p
                                                    key={"view"}
                                                    initial={{ opacity: 0, y: -10 }}
                                                    animate={{ opacity: 1, y: 0 }}
                                                    exit={{ opacity: 0, y: -10 }}
                                                    transition={{ duration: 0.2 }}
                                                    className={"text-md font-semibold px-2"}
                                                >
                                                    {routeDetails.data.max_idle_cons || "default"}
                                                </motion.p>
                                            )}
                                        </AnimatePresence>
                                    </div>
                                </div>
                                <div className={"flex flex-row w-full gap-5"}>
                                    <div className={"flex items-center justify-center"}>
                                        <div className={"bg-amber-500 p-2 rounded-full"}>
                                            <Antenna size={15}/>
                                        </div>
                                    </div>
                                    <div className={"flex flex-col gap-1 w-full"}>
                                        <p className={"text-md px-2"}>Max idle connections per host</p>
                                        <AnimatePresence mode={"wait"}>
                                            {editMode ? (
                                                <motion.div
                                                    key={"edit"}
                                                    initial={{ opacity: 0, y: -10 }}
                                                    animate={{ opacity: 1, y: 0 }}
                                                    exit={{ opacity: 0, y: -10 }}
                                                    transition={{ duration: 0.2 }}
                                                >
                                                    <Input
                                                        required
                                                        type={"number"}
                                                        value={editableRoute?.max_idle_cons_per_host || 0}
                                                        onChange={(e) => setEditableRoute(prev => prev ? {...prev, max_idle_cons_per_host: parseInt(e.target.value)} : null)}
                                                        className={`w-full px-3 py-2 rounded-lg bg-stone-700 text-white focus:outline-none focus:ring-2 focus:ring-amber-500`}
                                                    />
                                                </motion.div>
                                            ) : (
                                                <motion.p
                                                    key={"view"}
                                                    initial={{ opacity: 0, y: -10 }}
                                                    animate={{ opacity: 1, y: 0 }}
                                                    exit={{ opacity: 0, y: -10 }}
                                                    transition={{ duration: 0.2 }}
                                                    className={"text-md font-semibold px-2"}
                                                >
                                                    {routeDetails.data.max_idle_cons_per_host || "default"}
                                                </motion.p>
                                            )}
                                        </AnimatePresence>
                                    </div>
                                </div>
                                <div className={"flex flex-row w-full gap-5"}>
                                    <div className={"flex items-center justify-center"}>
                                        <div className={"bg-amber-500 p-2 rounded-full"}>
                                            <Hourglass size={15}/>
                                        </div>
                                    </div>
                                    <div className={"flex flex-col gap-1 w-full"}>
                                        <p className={"text-md px-2"}>Response header timeout</p>
                                        <AnimatePresence mode={"wait"}>
                                            {editMode ? (
                                                <motion.div
                                                    key={"edit"}
                                                    initial={{ opacity: 0, y: -10 }}
                                                    animate={{ opacity: 1, y: 0 }}
                                                    exit={{ opacity: 0, y: -10 }}
                                                    transition={{ duration: 0.2 }}
                                                >
                                                    <Input
                                                        required
                                                        type={"number"}
                                                        value={editableRoute?.response_header_timeout || 0}
                                                        onChange={(e) => setEditableRoute(prev => prev ? {...prev, response_header_timeout: parseInt(e.target.value)} : null)}
                                                        className={`w-full px-3 py-2 rounded-lg bg-stone-700 text-white focus:outline-none focus:ring-2 focus:ring-amber-500`}
                                                    />
                                                </motion.div>
                                            ) : (
                                                <motion.p
                                                    key={"view"}
                                                    initial={{ opacity: 0, y: -10 }}
                                                    animate={{ opacity: 1, y: 0 }}
                                                    exit={{ opacity: 0, y: -10 }}
                                                    transition={{ duration: 0.2 }}
                                                    className={"text-md font-semibold px-2"}
                                                >
                                                    {routeDetails.data.response_header_timeout || "default"}
                                                </motion.p>
                                            )}
                                        </AnimatePresence>
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
                                <p className={"text-xl font-semibold"}>Associated Plugins</p>
                                <div className={"flex items-center justify-center"}>
                                    <motion.button
                                        whileHover={{ scale: 1.05 }}
                                        whileTap={{ scale: 0.95 }}
                                        onClick={openNewPluginDialog}
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
                                                onClick={openNewPluginDialog}
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
                                                        <p className={"text-sm"}>Name</p>
                                                        <p className={"text-md font-semibold"}>{plugin.plugin?.name}</p>
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
                                        </div>
                                    )}
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
                    </div>
                </div>
            </div>
        </div>
    );
}

export default function RoutePage() {
    const links = [
        { label: "Routes", href: "/routes", active: false },
        { label: "Plugins", href: "/plugins", active: false },
        { label: "Settings", href: "/settings", active: false },
    ];

    return (
        <PageLayout>
            <NavBar links={links} />
            <RoutePageContent/>
        </PageLayout>
    );
}