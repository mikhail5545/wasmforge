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

import React, {useState, useEffect, useCallback} from "react";
import {useRouter, useSearchParams} from "next/navigation";
import {useData} from "@/hooks/useData";
import NavBar from "@/components/navigation/NavBar";
import {
    ListOrdered, CalendarClock,
    Trash, Pencil, BookKey,
    File, Route, Link2,
    ArrowUpRight, Check, X,
} from "lucide-react";
import {motion,AnimatePresence} from "motion/react";
import {JsonData, JsonEditor, monoDarkTheme} from "json-edit-react";
import PageLayout from "@/components/layout/PageLayout";
import {Input} from "@headlessui/react";
import {useMutation} from "@/hooks/useMutation";
import {ModalDialog} from "@/components/dialog/ModalDialog";
import Scrollbar from "react-scrollbars-custom";

function RoutePluginPageContent() {
    const router = useRouter();
    const params = useSearchParams();
    const pluginId = params.get("id");
    const routePluginDetails = useData<WasmForge.RoutePlugin>(`http://localhost:8080/api/route-plugins/${pluginId}`, "route_plugin");
    const [pluginDetails, setPluginDetails] = useState<{
        loading: boolean;
        data: WasmForge.Plugin | null;
        error: WasmForge.ErrorResponse | null;
    }>({loading: true, data: null, error: null});
    const [routeDetails, setRouteDetails] = useState<{
        loading: boolean;
        data: WasmForge.Route | null;
        error: WasmForge.ErrorResponse | null;
    }>({loading: true, data: null, error: null});

    const [showDeleteConfirmation, setShowDeleteConfirmation] = useState(false);
    const [editExecutionOrder, setEditExecutionOrder] = useState(false);
    const [editJSONConfig, setEditJSONConfig] = useState(false);
    const [jsonConfig, setJsonConfig] = useState<JsonData>({});
    const [routePluginEditableData, setRoutePluginEditableData] = useState<Omit<WasmForge.RoutePlugin, "id" | "created_at" | "route_id" | "plugin_id"> | null>(null);
    const mutation = useMutation();

    useEffect(() => {
        const plugin_id = routePluginDetails.data?.plugin_id;
        const route_id = routePluginDetails.data?.route_id;

        if (!plugin_id || !route_id) return;

        const controller = new AbortController();

        const fetchPluginDetails = async () => {
            setPluginDetails({loading: true, data: null, error: null});
            try{
                const res = await fetch(`http://localhost:8080/api/plugins/${plugin_id}`, { signal: controller.signal });
                if (!res.ok) {
                    const errorData = await res.json();
                    setPluginDetails({loading: false, data: null, error: errorData});
                } else {
                    const data = await res.json();
                    const pluginData = data["plugin"] as WasmForge.Plugin;
                    setPluginDetails({loading: false, data: pluginData, error: null});
                }
            } catch (err: any) {
                if (err.name === "AbortError") return;
                setPluginDetails({loading: false, data: null, error: {code: "", message: err.message, details: ""}});
            }
        };

        const fetchRouteDetails = async () => {
            setRouteDetails({loading: true, data: null, error: null});
            try{
                const res = await fetch(`http://localhost:8080/api/routes/${route_id}`, { signal: controller.signal });
                if (!res.ok) {
                    const errorData = await res.json();
                    setRouteDetails({loading: false, data: null, error: errorData});
                } else {
                    const data = await res.json();
                    const routeData = data["route"] as WasmForge.Route;
                    setRouteDetails({loading: false, data: routeData, error: null});
                }
            } catch (err: any) {
                if (err.name === "AbortError") return;
                setRouteDetails({loading: false, data: null, error: {code: "", message: err.message, details: ""}});
            }
        };

        fetchPluginDetails();
        fetchRouteDetails();
        return () => controller.abort();
    }, [routePluginDetails.data?.route_id, routePluginDetails.data?.plugin_id]);

    const handleSubmitChange = useCallback(
        async () => {
            if (!routePluginEditableData) return;

            try{
                const configString = JSON.stringify(jsonConfig);
                setRoutePluginEditableData(prev => prev ? ({ ...prev, config: configString }) : null);
            } catch (err: any) {
                mutation.setError({ code: "UNPROCESSABLE_ENTITY", "message": "Invalid JSON config", "details": err instanceof Error ? err.message : String(err)});
                console.error("Invalid JSON config:", err);
            }

            const res = await mutation.mutate(`http://localhost:8080/api/route-plugins/${pluginId}`, "PATCH", JSON.stringify(routePluginEditableData));
            if (res.success) {
                await routePluginDetails.refetch();
                setEditExecutionOrder(false);
                setEditJSONConfig(false);
                setRoutePluginEditableData(null);
            }
        }, [routePluginEditableData, jsonConfig, mutation, routePluginDetails, pluginId]
    );

    const deleteRoutePlugin = useCallback(
        async () => {
            const res = await mutation.mutate(`http://localhost:8080/api/route-plugins/${pluginId}`, "DELETE");
            if (res.success) {
                router.push("/plugins");
            }
        }, [mutation, pluginId, router]
    );

    const globalError = routePluginDetails.error || pluginDetails.error || routeDetails.error || mutation.error;
    const unsetError = async () => {
        mutation.setError(null);
        if (routePluginDetails.error) {
            await routePluginDetails.refetch();
        }
        if (pluginDetails.error) {
            router.refresh();
        }
        if (routeDetails.error) {
            router.refresh();
        }
    };

    return (
        <div className={"flex flex-col gap-5 w-full mt-20"}>
            <div className={"flex flex-row px-4 items-center justify-between bg-stone-800 rounded-4xl p-3 w-1/3"}>
                <p className={"text-xl font-semibold"}>Route Plugin Details</p>
                <div className={"flex flex-row"}>
                    <motion.button
                        whileHover={{ scale: 1.05 }}
                        whileTap={{ scale: 0.95 }}
                        disabled={mutation.loading}
                        onClick={() => setShowDeleteConfirmation(true)}
                        className={"px-2 py-2 rounded-full bg-red-500 text-white hover:bg-red-500/80 transition-colors duration-200 disabled:opacity-70 disabled:cursor-not-allowed"}
                    >
                        <Trash size={15}/>
                    </motion.button>
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
            <ModalDialog title={"Delete Route Plugin"} visible={showDeleteConfirmation} onClose={() => setShowDeleteConfirmation(false)} >
                <div className={"flex flex-col gap-5"}>
                    <p className={"text-md"}>Are you sure? This action is irreversible.</p>
                    <div className={"flex flex-row items-end justify-end"}>
                        <motion.button
                            whileHover={{ scale: 1.05 }}
                            whileTap={{ scale: 0.95 }}
                            onClick={deleteRoutePlugin}
                            className={"px-3 py-2 rounded-full bg-red-500 text-white hover:bg-red-500/80 transition-colors duration-200"}
                        >
                            Delete Route Plugin
                        </motion.button>
                    </div>
                </div>
            </ModalDialog>
            <div className={"flex flex-col lg:flex-row gap-5 w-full"}>
                <div className={"w-full lg:w-1/3 flex flex-col gap-5"}>
                    <div className={"border-box w-full bg-stone-800 rounded-4xl p-5"}>
                        {routePluginDetails.loading ? (
                            <div className={"flex items-center justify-center py-20"}>
                                <div className={"w-10 h-10 border-4 border-t-white border-stone-800 rounded-full animate-spin"}/>
                            </div>
                        ) : (
                            <div className={"flex flex-col gap-5"}>
                                <div className={"flex flex-row justify-between items-center"}>
                                    <div className={"flex flex-row w-full gap-5"}>
                                        <div className={"flex items-center justify-center"}>
                                            <div className={"bg-amber-500 p-2 rounded-full"}>
                                                <ListOrdered size={15}/>
                                            </div>
                                        </div>
                                        <div className={"flex flex-col gap-1"}>
                                            <p className={"text-md px-2"}>Execution Order</p>
                                            <AnimatePresence mode={"wait"}>
                                                {editExecutionOrder ? (
                                                    <motion.div
                                                        key={"edit-port"}
                                                        initial={{ opacity: 0, y: -10 }}
                                                        animate={{ opacity: 1, y: 0 }}
                                                        exit={{ opacity: 0, y: -10 }}
                                                        transition={{ duration: 0.2 }}
                                                    >
                                                        <Input
                                                            required
                                                            type={"number"}
                                                            value={routePluginEditableData?.execution_order || 0}
                                                            disabled={!editExecutionOrder}
                                                            onChange={(e) => setRoutePluginEditableData(prev => prev ? ({ ...prev, execution_order: parseInt(e.target.value) }) : null)}
                                                            className={`w-full px-3 py-2 rounded-lg bg-stone-700 text-white focus:outline-none focus:ring-2 focus:ring-amber-500`}
                                                        />
                                                    </motion.div>
                                                ) : (
                                                    <motion.p
                                                        key={"view-port"}
                                                        initial={{ opacity: 0, y: -10 }}
                                                        animate={{ opacity: 1, y: 0 }}
                                                        exit={{ opacity: 0, y: -10 }}
                                                        transition={{ duration: 0.2 }}
                                                        className={"text-md font-semibold px-2"}
                                                    >
                                                        {routePluginDetails.data.execution_order}
                                                    </motion.p>
                                                )}
                                            </AnimatePresence>
                                        </div>
                                    </div>
                                    <div className={"flex items-center justify-center gap-2"}>
                                        <AnimatePresence mode={"wait"}>
                                            {editExecutionOrder && (
                                                <motion.button
                                                    key={"submit-execution-order"}
                                                    whileHover={{ scale: 1.05 }}
                                                    whileTap={{ scale: 0.95 }}
                                                    disabled={mutation.loading || editJSONConfig}
                                                    onClick={handleSubmitChange}
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
                                            disabled={mutation.loading || editJSONConfig}
                                            onClick={editExecutionOrder ? () => {
                                                setEditExecutionOrder(false);
                                                setRoutePluginEditableData(routePluginDetails.data);
                                            } : () => {
                                                setEditExecutionOrder(true);
                                                setRoutePluginEditableData(routePluginDetails.data);
                                            }}
                                            className={"p-2 rounded-full  text-white hover:bg-white/5 transition-colors duration-200 disabled:opacity-70 disabled:cursor-not-allowed"}
                                        >
                                            {editExecutionOrder ? <X size={15}/> : <Pencil size={15}/> }
                                        </motion.button>
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
                                        <p className={"text-md font-semibold"}>{new Date(routePluginDetails.data.created_at).toLocaleString()}</p>
                                    </div>
                                </div>
                            </div>
                        )}
                    </div>
                    <div className={"border-box w-full bg-stone-800 rounded-4xl p-5"}>
                        {routePluginDetails.loading ? (
                            <div className={"flex items-center justify-center py-20"}>
                                <div className={"w-10 h-10 border-4 border-t-white border-stone-800 rounded-full animate-spin"}/>
                            </div>
                        ) : (
                            <div className={"flex flex-col gap-5"}>
                                <div className={"flex flex-row items-center justify-between"}>
                                    <p className={"text-lg font-semibold"}>JSON Config</p>
                                    <div className={"flex items-center justify-center gap-2"}>
                                        <AnimatePresence mode={"wait"}>
                                            {editJSONConfig && (
                                                <motion.button
                                                    key={"submit-config"}
                                                    whileHover={{ scale: 1.05 }}
                                                    whileTap={{ scale: 0.95 }}
                                                    disabled={mutation.loading || editExecutionOrder}
                                                    onClick={handleSubmitChange}
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
                                            disabled={mutation.loading || editExecutionOrder}
                                            onClick={editJSONConfig ? () => {
                                                setEditJSONConfig(false);
                                                setJsonConfig({});
                                            } : () => {
                                                setEditJSONConfig(true);
                                                try {
                                                    const parsedConfig = JSON.parse(routePluginDetails.data.config);
                                                    setJsonConfig(parsedConfig);
                                                } catch (err) {
                                                    console.error("Failed to parse JSON config:", err);
                                                    setJsonConfig({});
                                                }
                                            }}
                                            className={"p-2 rounded-full  text-white hover:bg-white/5 transition-colors duration-200 disabled:opacity-70 disabled:cursor-not-allowed"}
                                        >
                                            <Pencil size={15}/>
                                        </motion.button>
                                    </div>
                                </div>
                                <JsonEditor
                                    theme={monoDarkTheme}
                                    restrictEdit={!editJSONConfig}
                                    restrictAdd={!editJSONConfig}
                                    restrictDrag={!editJSONConfig}
                                    restrictDelete={!editJSONConfig}
                                    restrictTypeSelection={!editJSONConfig}
                                    setData={editJSONConfig ? (data) => setJsonConfig(data) : undefined}
                                    data={editJSONConfig ? jsonConfig : JSON.parse(routePluginDetails.data.config)}
                                />
                            </div>
                        )}
                    </div>
                </div>
                <div className={"w-full lg:w-1/3"}>
                    <div className={"border-box w-full bg-stone-800 rounded-4xl p-5"}>
                        {pluginDetails.loading ? (
                            <div className={"flex items-center justify-center py-20"}>
                                <div className={"w-10 h-10 border-4 border-t-white border-stone-800 rounded-full animate-spin"}/>
                            </div>
                        ) : (
                            <div>
                                {pluginDetails.data ? (
                                    <div className={"flex flex-col gap-5"}>
                                        <div className={"flex flex-row items-center justify-between"}>
                                            <p className={"text-lg font-semibold"}>Plugin Origin</p>
                                            <div className={"flex items-center justify-center"}>
                                                <motion.a
                                                    whileHover={{ scale: 1.05 }}
                                                    whileTap={{ scale: 0.95 }}
                                                    href={`/plugins/plugin?name=${pluginDetails.data.name}`}
                                                    className={"p-2 rounded-full bg-white text-black hover:bg-white/80 transition-colors duration-200"}
                                                >
                                                    <ArrowUpRight size={15}/>
                                                </motion.a>
                                            </div>
                                        </div>
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
                                                    <CalendarClock size={15}/>
                                                </div>
                                            </div>
                                            <div className={"flex flex-col"}>
                                                <p className={"text-md"}>Created at</p>
                                                <p className={"text-md font-semibold"}>{new Date(pluginDetails.data.created_at).toLocaleString()}</p>
                                            </div>
                                        </div>
                                    </div>
                                ) : (
                                    <div className={"flex items-center justify-center py-10"}>
                                        <p className={"text-lg font-semibold"}>Failed to fetch plugin details</p>
                                    </div>
                                )}
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
                            <div>
                                {routeDetails.data ? (
                                    <div className={"flex flex-col gap-5"}>
                                        <div className={"flex flex-row items-center justify-between"}>
                                            <p className={"text-lg font-semibold"}>Attached to Route</p>
                                            <div className={"flex items-center justify-center"}>
                                                <motion.a
                                                    whileHover={{ scale: 1.05 }}
                                                    whileTap={{ scale: 0.95 }}
                                                    href={`/routes/route?path=${routeDetails.data.path}`}
                                                    className={"p-2 rounded-full bg-white text-black hover:bg-white/80 transition-colors duration-200"}
                                                >
                                                    <ArrowUpRight size={15}/>
                                                </motion.a>
                                            </div>
                                        </div>
                                        <div className={"flex flex-row w-full gap-5"}>
                                            <div className={"flex items-center justify-center"}>
                                                <div className={"bg-amber-500 p-2 rounded-full"}>
                                                    <Route size={15}/>
                                                </div>
                                            </div>
                                            <div className={"flex flex-col"}>
                                                <p className={"text-md"}>Path</p>
                                                <p className={"text-md font-semibold"}>{routeDetails.data.path}</p>
                                            </div>
                                        </div>
                                        <div className={"flex flex-row w-full gap-5"}>
                                            <div className={"flex items-center justify-center"}>
                                                <div className={"bg-amber-500 p-2 rounded-full"}>
                                                    <Link2 size={15}/>
                                                </div>
                                            </div>
                                            <div className={"flex flex-col"}>
                                                <p className={"text-md"}>Target URL</p>
                                                <p className={"text-md font-semibold"}>{routeDetails.data.target_url}</p>
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
                                                <p className={"text-md font-semibold"}>{new Date(routeDetails.data.created_at).toLocaleString()}</p>
                                            </div>
                                        </div>
                                    </div>
                                ) : (
                                    <div className={"flex items-center justify-center py-10"}>
                                        <p className={"text-lg font-semibold"}>Failed to fetch route details</p>
                                    </div>
                                )}
                            </div>
                        )}
                    </div>
                </div>
            </div>
        </div>
    );
}

export default function RoutePluginPage() {
    const links = [
        { label: "Routes", href: "/routes", active: false },
        { label: "Plugins", href: "/plugins", active: false },
        { label: "Settings", href: "/settings", active: false },
    ];

    return (
        <PageLayout>
            <NavBar links={links} />
            <RoutePluginPageContent/>
        </PageLayout>
    );
}