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

import React, {useState, useEffect, useCallback, Suspense} from "react";
import NavBar from "@/components/navigation/NavBar";
import {useSearchParams,useRouter} from "next/navigation";
import {useData} from "@/hooks/useData";
import {motion, AnimatePresence} from "motion/react";
import {Pencil, X, ChevronLeft, ChevronRight} from "lucide-react";
import {usePaginatedData} from "@/hooks/usePaginatedData";
import {PluginGridListCard,RouteGridListCard} from "@/components/card/GridListCard";
import {JsonEditor,githubDarkTheme,JsonData} from "json-edit-react";
import {
    Fieldset,
    Label,
    Legend,
    Field,
    Input,
} from "@headlessui/react";
import {useMutation} from "@/hooks/useMutation";
import {InfoDialog} from "@/components/dialog/InfoDialog";
import {ErrorDialog} from "@/components/dialog/ErrorDialog";

const initialRoutePluginFormState: Omit<WasmForge.RoutePlugin, "id" | "created_at" | "plugin"> = {
    route_id: "",
    plugin_id: "",
    execution_order: 1,
    config: "",
};

function NewRoutePluginPageContent() {
   const links = [
        { label: "Dashboard", href: "/dashboard" },
        { label: "Routes", href: "/routes" },
        { label: "Plugins", href: "/plugins" },
   ];

    const searchParams = useSearchParams();
    const routeId = searchParams.get("route_id");
    const pluginId = searchParams.get("plugin_id");

    const [selectedRoute, setSelectedRoute] = useState<WasmForge.Route | null>(null);
    const [selectedPlugin, setSelectedPlugin] = useState<WasmForge.Plugin | null>(null);
    const [showRouteSelector, setShowRouteSelector] = useState(false);
    const [showPluginSelector, setShowPluginSelector] = useState(false);
    const [routePluginFormState, setRoutePluginFormState] = useState<Omit<WasmForge.RoutePlugin, "id" | "created_at" | "plugin">>(initialRoutePluginFormState);
    const [jsonConfig, setJsonConfig] = useState<JsonData>({});
    const [success, setSuccess] = useState(false);
    const [globalError, setGlobalError] = useState<WasmForge.ErrorResponse | null>(null);

    // Call hooks unconditionally with nullable path arguments
    const routePath = routeId ? `http://localhost:8080/api/routes/${routeId}` : null;
    const pluginPath = pluginId ? `http://localhost:8080/api/plugins/${pluginId}` : null;

    const routeData = useData<WasmForge.Route>(routePath, "route");
    const pluginData = useData<WasmForge.Plugin>(pluginPath, "plugin");

    const mutation = useMutation();
    const router = useRouter();

    const paginatedPluginData = usePaginatedData<WasmForge.Plugin>(
        "http://localhost:8080/api/plugins",
        "plugins",
        10,
        "created_at",
        "desc",
    );

    const paginatedRouteData = usePaginatedData<WasmForge.Route>(
        "http://localhost:8080/api/routes",
        "routes",
        10,
        "created_at",
        "desc",
    );

    // Update selectedRoute when routeData changes (or clear when no routeId)
    useEffect(() => {
        if (!routeId) {
            setSelectedRoute(null);
            return;
        }

        if (routeData.error) {
            setGlobalError(routeData.error);
            setSelectedRoute(null);
        } else if (!routeData.loading && routeData.data) {
            setSelectedRoute(routeData.data);
            setRoutePluginFormState(prev => ({ ...prev, route_id: routeData.data.id}));
        }
    }, [routeId, routeData.data, routeData.loading, routeData.error]);

    // Update selectedPlugin when pluginData changes (or clear when no pluginId)
    useEffect(() => {
        if (!pluginId) {
            setSelectedPlugin(null);
            return;
        }

        if (pluginData.error) {
            setGlobalError(pluginData.error);
            setSelectedPlugin(null);
        } else if (!pluginData.loading && pluginData.data) {
            setSelectedPlugin(pluginData.data);
            setRoutePluginFormState(prev => ({ ...prev, plugin_id: pluginData.data.id}));
        }
    }, [pluginId, pluginData.data, pluginData.loading, pluginData.error]);

    const handleSubmitConfig = () => {
        // Validate JSON config before submitting
        try {
            const configString = JSON.stringify(jsonConfig);
            setRoutePluginFormState(prev => ({ ...prev, config: configString }));}
        catch (error) {
            setGlobalError({ code: "UNPROCESSABLE_ENTITY", "message": "Invalid JSON config", "details": error instanceof Error ? error.message : String(error)});
            console.error("Invalid JSON config:", error);
        }
    };

    const handleSubmit = useCallback(
        async() => {
            if (!selectedPlugin || !selectedRoute) {
                return;
            }

            const result = await mutation.mutate("http://localhost:8080/api/route-plugins", "POST", JSON.stringify(routePluginFormState));
            if (result.success){
                setSuccess(true);
            } else {
                setGlobalError(mutation.error);
            }

        }, [selectedPlugin, selectedRoute]
    );

    return (
        <div className={"flex flex-col w-full"}>
            <NavBar
                title={"WasmForge"}
                links={links}
            />
            <InfoDialog
                title={"Success"}
                message={"Plugin successfully created"}
                isOpen={success}
                onClose={() => router.push(selectedRoute ? `/routes?path=${selectedRoute.path}` : "/routes")}
            />
            <ErrorDialog
                title={globalError?.message || "Unexpected error"}
                message={globalError?.details || "No additional details available"}
                isOpen={!!globalError} onClose={() => setGlobalError(null)}
            />
            <div className={"py-10 px-5 md:px-15 lg:px-30"}>
                <h1 className={"text-xl font-semibold py-5"}>Add plugin to route</h1>
                <div className={"flex flex-col lg:flex-row gap-5"}>
                    <Fieldset className={"w-full lg:w-1/2 space-y-6 rounded-xl bg-white/5 p-6 sm:p-8"}>
                        <Legend className={"text-lg font-semibold text-white"}>
                            Plugin details
                            <p className={"text-sm font-semibold text-white/50"}>Required fields are marked with *</p>
                        </Legend>
                        <Field>
                            <Label className={"text-base/7 font-semibold text-white"}>Route ID</Label>
                            <Input
                                type={"text"}
                                value={routePluginFormState.route_id}
                                disabled={true}
                                className={"mt-1 w-full rounded-md border border-stone-700 bg-stone-800/50 px-3 py-2 text-sm text-white disabled:opacity-50 focus:ring focus:ring-white"}
                            />
                        </Field>
                        <Field>
                            <Label className={"text-base/7 font-semibold text-white"}>Plugin ID</Label>
                            <Input
                                type={"text"}
                                value={routePluginFormState.plugin_id}
                                disabled={true}
                                className={"mt-1 w-full rounded-md border border-stone-700 bg-stone-800/50 px-3 py-2 text-sm text-white disabled:opacity-50 focus:ring focus:ring-white"}
                            />
                        </Field>
                        <Field>
                            <Label className={"text-base/7 font-semibold text-white"}>Execution Order*</Label>
                            <Input
                                required
                                type={"number"}
                                value={routePluginFormState.execution_order}
                                onChange={(e) => setRoutePluginFormState(prev => ({ ...prev, execution_order: parseInt(e.target.value) }))}
                                className={"mt-1 w-full rounded-md border border-stone-700 bg-stone-800/50 px-3 py-2 text-sm text-white  focus:ring focus:ring-white"}
                            />
                        </Field>
                        <div className={"flex flex-row gap-5 pt-4 justify-start"}>
                            <motion.button
                                whileHover={{ scale: 1.05 }}
                                whileTap={{ scale: 0.95 }}
                                disabled={!selectedPlugin || !selectedRoute}
                                onClick={handleSubmit}
                                className={"bg-stone-800 text-sm font-semibold px-3 py-1 rounded disabled:opacity-50 flex items-center justify-center gap-2"}
                            >
                                {mutation.loading ? (
                                    <div className={"w-5 h-5 border-3 border-t-stone-800 border-white rounded-full animate-spin"}/>
                                ) : (
                                    <p>Submit</p>
                                )}
                            </motion.button>
                        </div>
                    </Fieldset>
                    <Fieldset className={"w-full lg:w-1/2 rounded-xl space-y-6 bg-white/5 p-6 sm:p-8"}>
                        <div className={"flex flex-col justify-between h-full"}>
                            <div className={"space-y-6"}>
                                <Legend className={"text-lg font-semibold text-white"}>
                                    Custom JSON configuration
                                    <p className={"text-sm font-semibold text-white/50"}>You can specify JSON config that plugins can access during runtime</p>
                                </Legend>
                                <Field>
                                    <JsonEditor
                                        className={"w-full"}
                                        theme={githubDarkTheme}
                                        data={jsonConfig} setData={(data) => {
                                        setJsonConfig(data);
                                    }}/>
                                </Field>
                            </div>
                            <div className={"flex flex-row gap-5 pt-4 justify-start"}>
                                <motion.button
                                    whileHover={{ scale: 1.05 }}
                                    whileTap={{ scale: 0.95 }}
                                    disabled={!selectedPlugin || !selectedRoute}
                                    onClick={handleSubmitConfig}
                                    className={"bg-stone-800 text-sm font-semibold px-3 py-1 rounded disabled:opacity-50 flex items-center justify-center gap-2"}
                                >
                                    Submit
                                </motion.button>
                            </div>
                        </div>
                    </Fieldset>
                </div>
                <div className={"flex flex-col lg:flex-row gap-5 mt-5"}>
                    <div className={"flex w-full lg:w-1/2"}>
                        <div className={"border-box w-full border border-stone-700 rounded"}>
                            <div className={"flex flex-col space-y-6 p-5"}>
                                {!showPluginSelector ? (
                                    <div className={"flex flex-row gap-5"}>
                                        <h2 className={"text-lg font-semibold"}>
                                            Selected Plugin
                                        </h2>
                                        <motion.button
                                            whileHover={{ scale: 1.1 }}
                                            whileTap={{ scale: 0.9 }}
                                            onClick={ async () => {setShowPluginSelector(true); await paginatedPluginData.refetch()} }
                                            className={"px-2 py-2 bg-stone-800 text-sm rounded flex items-center justify-center text-center"}
                                        >
                                            <Pencil size={10}/>
                                        </motion.button>
                                    </div>
                                ) : (
                                    <div className={"flex flex-row gap-5"}>
                                        <h2 className={"text-lg font-semibold"}>
                                            Select a Plugin
                                        </h2>
                                        <motion.button
                                            whileHover={{ scale: 1.1 }}
                                            whileTap={{ scale: 0.9 }}
                                            onClick={ () => setShowPluginSelector(false) }
                                            className={"px-2 py-2 bg-stone-800 text-sm rounded flex items-center justify-center text-center"}
                                        >
                                            <X size={10}/>
                                        </motion.button>
                                    </div>
                                )}
                                <AnimatePresence
                                    mode={"wait"}
                                >
                                    {!selectedPlugin && !showPluginSelector && (
                                        <motion.div
                                            key={"selected-plugin-info"}
                                            initial={{ opacity: 0, x: -20 }}
                                            animate={{ opacity: 1, x: 0 }}
                                            exit={{ opacity: 0, x: 20 }}
                                            transition={{ duration: 0.3 }}
                                            className={"flex flex-col gap-4 p-5 bg-stone-800 rounded w-full"}
                                        >
                                            <div className={"text-start"}>
                                                <p className={"text-lg font-semibold"}>No plugin selected</p>
                                                <p className={"text-md text-stone-400"}>Click the edit button to select a plugin to add to this route.</p>
                                            </div>
                                        </motion.div>
                                    )}
                                    {selectedPlugin && !showPluginSelector &&(
                                        <motion.div
                                            key={"selected-plugin-info"}
                                            initial={{ opacity: 0, x: -20 }}
                                            animate={{ opacity: 1, x: 0 }}
                                            exit={{ opacity: 0, x: 20 }}
                                            transition={{ duration: 0.3 }}
                                            className={"flex flex-col gap-4 p-5 bg-stone-800 rounded w-full"}
                                        >
                                            <div className={"flex flex-row justify-between"}>
                                                <p className={"text-md text-stone-400"}>Name</p>
                                                <p className={"text-md font-semibold"}>{selectedPlugin.name}</p>
                                            </div>
                                            <div className={"flex flex-row justify-between"}>
                                                <p className={"text-md text-stone-400"}>Filename</p>
                                                <p className={"text-md font-semibold"}>{selectedPlugin.filename}</p>
                                            </div>
                                        </motion.div>
                                    )}
                                    {showPluginSelector && (
                                        <motion.div
                                            key={"plugin-selector"}
                                            initial={{ opacity: 0, x: -20 }}
                                            animate={{ opacity: 1, x: 0 }}
                                            exit={{ opacity: 0, x: 20 }}
                                            transition={{ duration: 0.3 }}
                                            className={"flex flex-col gap-4 p-5  rounded w-full"}
                                        >
                                            {paginatedPluginData.loading ? (
                                                <div className={"flex justify-center items-center py-20"}>
                                                    <div className={"w-10 h-10 border-4 border-t-white border-stone-600 rounded-full animate-spin"}/>
                                                </div>
                                            ) : (
                                                <div>
                                                    {paginatedPluginData.data.length === 0 ? (
                                                        <div className={"flex justify-center items-center py-20"}>
                                                            <p className={"text-center text-stone-400"}>No plugins found.</p>
                                                        </div>
                                                    ) : (
                                                        <>
                                                            <div className={"grid grid-cols-1 gap-1"}>
                                                                {paginatedPluginData.data.map((plugin, idx) => (
                                                                    <PluginGridListCard
                                                                        key={plugin.id}
                                                                        plugin={plugin}
                                                                        index={idx}
                                                                        onClick={() => setSelectedPlugin(plugin)}
                                                                        currentlySelected={plugin.id === selectedPlugin?.id}
                                                                    />
                                                                ))}
                                                            </div>
                                                            <div className={"flex flex-row justify-between items-center mt-5"}>
                                                                <motion.button
                                                                    whileHover={{ scale: 1.05 }}
                                                                    whileTap={{ scale: 0.95 }}
                                                                    onClick={ async () => await paginatedPluginData.refetch() }
                                                                    className={"border border-stone-800 text-sm font-semibold px-3 py-1 rounded disabled:opacity-50 flex items-center justify-center gap-2"}
                                                                >
                                                                    <ChevronLeft size={10}/>First page
                                                                </motion.button>
                                                                <motion.button
                                                                    whileHover={{ scale: 1.05 }}
                                                                    whileTap={{ scale: 0.95 }}
                                                                    disabled={paginatedPluginData.nextPageToken === ""}
                                                                    onClick={ async () => { await paginatedPluginData.nextPage(paginatedPluginData.nextPageToken, { append: false })} }
                                                                    className={"border border-stone-800 text-sm font-semibold px-3 py-1 rounded disabled:cursor-not-allowed disabled:opacity-50 flex items-center justify-center gap-2"}
                                                                >
                                                                    Next page<ChevronRight size={10}/>
                                                                </motion.button>
                                                            </div>
                                                            <div className={"flex flex-row justify-between items-center mt-5"}>
                                                                <motion.button
                                                                    whileHover={{ scale: 1.05 }}
                                                                    whileTap={{ scale: 0.95 }}
                                                                    onClick={ () => setShowPluginSelector(false) }
                                                                    className={"border border-stone-800 text-sm font-semibold px-3 py-1 rounded disabled:opacity-50 flex items-center justify-center gap-2"}
                                                                >
                                                                    <ChevronLeft size={10}/>Back
                                                                </motion.button>
                                                                <motion.button
                                                                    whileHover={{ scale: 1.05 }}
                                                                    whileTap={{ scale: 0.95 }}
                                                                    disabled={!selectedPlugin}
                                                                    onClick={ () => setShowPluginSelector(false) }
                                                                    className={"bg-stone-800 text-sm font-semibold px-3 py-1 rounded disabled:opacity-50 flex items-center justify-center gap-2"}
                                                                >
                                                                    Submit selection<ChevronRight size={10}/>
                                                                </motion.button>
                                                            </div>
                                                        </>
                                                    )}
                                                </div>
                                            )}
                                        </motion.div>
                                    )}
                                </AnimatePresence>
                            </div>
                        </div>
                    </div>
                    <div className={"flex w-full lg:w-1/2"}>
                        <div className={"border-box w-full border border-stone-700 rounded"}>
                            <div className={"flex flex-col space-y-6 p-5"}>
                                {!showRouteSelector ? (
                                    <div className={"flex flex-row gap-5"}>
                                        <h2 className={"text-lg font-semibold"}>
                                            Selected Route
                                        </h2>
                                        <motion.button
                                            whileHover={{ scale: 1.1 }}
                                            whileTap={{ scale: 0.9 }}
                                            onClick={ async () => {setShowRouteSelector(true); await paginatedRouteData.refetch()} }
                                            className={"px-2 py-2 bg-stone-800 text-sm rounded flex items-center justify-center text-center"}
                                        >
                                            <Pencil size={10}/>
                                        </motion.button>
                                    </div>
                                ) : (
                                    <div className={"flex flex-row gap-5"}>
                                        <h2 className={"text-lg font-semibold"}>
                                            Select a Route
                                        </h2>
                                        <motion.button
                                            whileHover={{ scale: 1.1 }}
                                            whileTap={{ scale: 0.9 }}
                                            onClick={ () => setShowRouteSelector(false) }
                                            className={"px-2 py-2 bg-stone-800 text-sm rounded flex items-center justify-center text-center"}
                                        >
                                            <X size={10}/>
                                        </motion.button>
                                    </div>
                                )}
                                <AnimatePresence
                                    mode={"wait"}
                                >
                                    {!selectedRoute && !showRouteSelector && (
                                        <motion.div
                                            key={"selected-plugin-info"}
                                            initial={{ opacity: 0, x: -20 }}
                                            animate={{ opacity: 1, x: 0 }}
                                            exit={{ opacity: 0, x: 20 }}
                                            transition={{ duration: 0.3 }}
                                            className={"flex flex-col gap-4 p-5 bg-stone-800 rounded w-full"}
                                        >
                                            <div className={"text-start"}>
                                                <p className={"text-lg font-semibold"}>No route selected</p>
                                                <p className={"text-md text-stone-400"}>Click the edit button to select a route.</p>
                                            </div>
                                        </motion.div>
                                    )}
                                    {selectedRoute && !showRouteSelector && (
                                        <motion.div
                                            key={"selected-plugin-info"}
                                            initial={{ opacity: 0, x: -20 }}
                                            animate={{ opacity: 1, x: 0 }}
                                            exit={{ opacity: 0, x: 20 }}
                                            transition={{ duration: 0.3 }}
                                            className={"flex flex-col gap-4 p-5 bg-stone-800 rounded w-full"}
                                        >
                                            <div className={"flex flex-row justify-between"}>
                                                <p className={"text-md text-stone-400"}>Path</p>
                                                <p className={"text-md font-semibold"}>{selectedRoute.path}</p>
                                            </div>
                                            <div className={"flex flex-row justify-between"}>
                                                <p className={"text-md text-stone-400"}>Target URL</p>
                                                <p className={"text-md font-semibold"}>{selectedRoute.target_url}</p>
                                            </div>
                                        </motion.div>
                                    )}
                                    {showRouteSelector && (
                                        <motion.div
                                            key={"plugin-selector"}
                                            initial={{ opacity: 0, x: -20 }}
                                            animate={{ opacity: 1, x: 0 }}
                                            exit={{ opacity: 0, x: 20 }}
                                            transition={{ duration: 0.3 }}
                                            className={"flex flex-col gap-4 p-5  rounded w-full"}
                                        >
                                            {paginatedRouteData.loading ? (
                                                <div className={"flex justify-center items-center py-20"}>
                                                    <div className={"w-10 h-10 border-4 border-t-white border-stone-600 rounded-full animate-spin"}/>
                                                </div>
                                            ) : (
                                                <div>
                                                    {paginatedRouteData.data.length === 0 ? (
                                                        <div className={"flex justify-center items-center py-20"}>
                                                            <p className={"text-center text-stone-400"}>No plugins found.</p>
                                                        </div>
                                                    ) : (
                                                        <>
                                                            <div className={"grid grid-cols-1 gap-1"}>
                                                                {paginatedRouteData.data.map((route, idx) => (
                                                                    <RouteGridListCard
                                                                        key={route.id}
                                                                        route={route}
                                                                        index={idx}
                                                                        onClick={() => setSelectedRoute(route)}
                                                                        currentlySelected={route.id === selectedRoute?.id}
                                                                    />
                                                                ))}
                                                            </div>
                                                            <div className={"flex flex-row justify-between items-center mt-5"}>
                                                                <motion.button
                                                                    whileHover={{ scale: 1.05 }}
                                                                    whileTap={{ scale: 0.95 }}
                                                                    onClick={ async () => await paginatedRouteData.refetch() }
                                                                    className={"border border-stone-800 text-sm font-semibold px-3 py-1 rounded disabled:opacity-50 flex items-center justify-center gap-2"}
                                                                >
                                                                    <ChevronLeft size={10}/>First page
                                                                </motion.button>
                                                                <motion.button
                                                                    whileHover={{ scale: 1.05 }}
                                                                    whileTap={{ scale: 0.95 }}
                                                                    disabled={paginatedRouteData.nextPageToken === ""}
                                                                    onClick={ async () => { await paginatedRouteData.nextPage(paginatedRouteData.nextPageToken, { append: false })} }
                                                                    className={"border border-stone-800 text-sm font-semibold px-3 py-1 rounded disabled:cursor-not-allowed disabled:opacity-50 flex items-center justify-center gap-2"}
                                                                >
                                                                    Next page<ChevronRight size={10}/>
                                                                </motion.button>
                                                            </div>
                                                            <div className={"flex flex-row justify-between items-center mt-5"}>
                                                                <motion.button
                                                                    whileHover={{ scale: 1.05 }}
                                                                    whileTap={{ scale: 0.95 }}
                                                                    onClick={ () => setShowRouteSelector(false) }
                                                                    className={"border border-stone-800 text-sm font-semibold px-3 py-1 rounded disabled:opacity-50 flex items-center justify-center gap-2"}
                                                                >
                                                                    <ChevronLeft size={10}/>Back
                                                                </motion.button>
                                                                <motion.button
                                                                    whileHover={{ scale: 1.05 }}
                                                                    whileTap={{ scale: 0.95 }}
                                                                    disabled={!selectedPlugin}
                                                                    onClick={ () => setShowRouteSelector(false) }
                                                                    className={"bg-stone-800 text-sm font-semibold px-3 py-1 rounded disabled:opacity-50 flex items-center justify-center gap-2"}
                                                                >
                                                                    Submit selection<ChevronRight size={10}/>
                                                                </motion.button>
                                                            </div>
                                                        </>
                                                    )}
                                                </div>
                                            )}
                                        </motion.div>
                                    )}
                                </AnimatePresence>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
}

export default function NewRoutePluginPage() {
    return(
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
                <NewRoutePluginPageContent />
            </div>
        </Suspense>
    );
}