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
import NavBar from "@/components/navigation/NavBar";
import {useSearchParams,useRouter} from "next/navigation";
import {useData} from "@/hooks/useData";
import {usePaginatedData} from "@/hooks/usePaginatedData";
import {JsonEditor,githubDarkTheme,JsonData} from "json-edit-react";
import PageLayout from "@/components/layout/PageLayout";
import {useMutation} from "@/hooks/useMutation";
import {motion} from "motion/react";
import {ArrowLeft, Undo2, Link2, Route, BookKey, File, Pencil, Plus} from "lucide-react";
import {ModalDialog} from "@/components/dialog/ModalDialog";
import {Scrollbar} from "react-scrollbars-custom";
import {Input} from "@headlessui/react";

const initialRoutePluginFormState: Omit<WasmForge.RoutePlugin, "id" | "created_at" | "plugin"> = {
    route_id: "",
    plugin_id: "",
    execution_order: 1,
    config: "",
};

export default function NewRoutePluginPage() {
    const links = [
        { label: "Routes", href: "/routes", active: false },
        { label: "Plugins", href: "/plugins", active: false },
        { label: "Settings", href: "/settings", active: false },
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

    const handleSubmit = useCallback(
        async() => {
            if (!selectedPlugin || !selectedRoute) {
                return;
            }

            try {
                const configString = JSON.stringify(jsonConfig);
                setRoutePluginFormState(prev => ({ ...prev, config: configString }));}
            catch (error) {
                setGlobalError({ code: "UNPROCESSABLE_ENTITY", "message": "Invalid JSON config", "details": error instanceof Error ? error.message : String(error)});
                console.error("Invalid JSON config:", error);
            }

            const result = await mutation.mutate("http://localhost:8080/api/route-plugins", "POST", JSON.stringify(routePluginFormState));
            if (result.success){
                setSuccess(true);
            } else {
                setGlobalError(mutation.error);
            }

        }, [selectedPlugin, selectedRoute]
    );

    return(
        <PageLayout>
            <NavBar links={links} />
            <div className={"flex flex-col gap-5 w-full mt-20"}>
                <div className={"flex flex-row px-4 items-center justify-between bg-stone-800 rounded-4xl p-3 w-1/3"}>
                    <p className={"text-xl font-semibold"}>Creating a new Route Plugin</p>
                    <div className={"flex flex-row gap-2"}>
                        <motion.button
                            whileHover={{ scale: 1.05 }}
                            whileTap={{ scale: 0.95 }}
                            onClick={() => router.back()}
                            className={"px-2 py-2 rounded-full bg-white text-black hover:bg-white/80 transition-colors duration-200"}
                        >
                            <ArrowLeft size={15}/>
                        </motion.button>
                        <motion.button
                            whileHover={{ scale: 1.05 }}
                            whileTap={{ scale: 0.95 }}
                            className={"px-2 py-2 rounded-full bg-amber-500 text-white hover:bg-amber-500/80 transition-colors duration-200"}
                        >
                            <Undo2 size={15}/>
                        </motion.button>
                    </div>
                </div>
            </div>
            <ModalDialog title={"Route Plugin successfully created"} visible={success} onClose={() => router.push(`/routes/route?path=${selectedRoute?.path}`)} >
                <div className={"flex flex-col gap-5"}>
                    <p className={"text-md font-semibold"}>You successfully created a new route plugin!</p>
                    <p className={"text-sm"}>When route is enabled, it will be applied as middleware according to execution order</p>
                    <div className={"flex flex-row items-end justify-end"}>
                        <motion.button
                            whileHover={{ scale: 1.05 }}
                            whileTap={{ scale: 0.95 }}
                            onClick={() => router.push(`/routes/route?path=${selectedRoute?.path}`)}
                            className={"px-3 py-2 rounded-full bg-white text-black hover:bg-white/80 transition-colors duration-200"}
                        >
                            Close
                        </motion.button>
                    </div>
                </div>
            </ModalDialog>
            <ModalDialog title={"Error"} visible={!!globalError} onClose={() => setGlobalError(null)} >
                <div className={"flex flex-col gap-5"}>
                    <p className={"text-md font-semibold"}>{globalError?.message}</p>
                    <Scrollbar style={{ height: 200 }}>
                        <p className={"text-md"}>{globalError?.details}</p>
                    </Scrollbar>
                    <div className={"flex flex-row items-end justify-end"}>
                        <motion.button
                            whileHover={{ scale: 1.05 }}
                            whileTap={{ scale: 0.95 }}
                            onClick={() => setGlobalError(null)}
                            className={"px-3 py-2 rounded-full bg-white text-black hover:bg-white/80 transition-colors duration-200"}
                        >
                            Close
                        </motion.button>
                    </div>
                </div>
            </ModalDialog>
            <ModalDialog title={"Select a Route"} visible={showRouteSelector} onClose={() => setShowRouteSelector(false)}>
                <div className={"flex flex-col gap-5"}>
                    {paginatedRouteData.loading ? (
                        <div className={"flex items-center justify-center py-10"}>
                            <div className={"w-10 h-10 border-4 border-t-white border-stone-800 rounded-full animate-spin"}/>
                        </div>
                    ) : (
                        <div>
                            {paginatedRouteData.data.length === 0 ? (
                                <div className={"flex flex-col items-center justify-center py-10"}>
                                    <p className={"text-lg"}>You didn't create any routes yet</p>
                                    <a href={"/routes/new"} className={"text-lg underline"}>Start with creating one</a>
                                </div>
                            ) : (
                                <Scrollbar style={{ height: 500 }}>
                                    <div className={"grid grid-cols-1 gap-2 pr-2 mt-5"}>
                                        {paginatedRouteData.data.map((route, idx) => (
                                            <motion.div
                                                key={route.id}
                                                initial={{ opacity: 0, y: 10 }}
                                                animate={{ opacity: 1, y: 0 }}
                                                transition={{ duration: 0.2, delay: idx * 0.1 }}
                                                onClick={() => setSelectedRoute(route)}
                                                className={`px-4 col-span-1 flex flex-row items-center p-1 rounded-2xl gap-20 ${
                                                    selectedRoute?.id === route.id ? "bg-amber-500" : "border border-amber-500 hover:bg-amber-500 transition-colors duration-200"
                                                }`}
                                            >
                                                <div className={"flex flex-col items-start justify-center"}>
                                                    <p className={"text-sm"}>Path</p>
                                                    <p className={"text-md font-semibold truncate"}>{route.path}</p>
                                                </div>
                                                <div className={"flex flex-col items-start justify-center"}>
                                                    <p className={"text-sm"}>Target URL</p>
                                                    <p className={"text-md font-semibold truncate"}>{route.target_url}</p>
                                                </div>
                                            </motion.div>
                                        ))}
                                    </div>
                                </Scrollbar>
                            )}
                            {paginatedRouteData.nextPageToken && (
                                <div className={"flex items-center justify-center pt-3"}>
                                    <motion.button
                                        whileHover={{ scale: 1.05 }}
                                        whileTap={{ scale: 0.95 }}
                                        onClick={() => paginatedRouteData.nextPage(paginatedRouteData.nextPageToken, {append: true, force: true})}
                                        className={"text-sm underline"}
                                    >
                                        Load more
                                    </motion.button>
                                </div>
                            )}
                            <div className={"flex flex-row items-center justify-end pt-5"}>
                                <div className={"flex flex-row items-center justify-between gap-5"}>
                                    <motion.button
                                        whileHover={{ scale: 1.05 }}
                                        whileTap={{ scale: 0.95 }}
                                        onClick={() => setShowRouteSelector(false)}
                                        className={"px-3 py-2 rounded-full bg-white text-black hover:bg-white/80 transition-colors duration-200 disabled:opacity-70 disabled:cursor-not-allowed"}
                                    >
                                        Cancel
                                    </motion.button>
                                    <motion.button
                                        disabled={!selectedRoute}
                                        whileHover={{ scale: 1.05 }}
                                        whileTap={{ scale: 0.95 }}
                                        onClick={() => setShowRouteSelector(false)}
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
            <ModalDialog title={"Select a Plugin"} visible={showPluginSelector} onClose={() => setShowPluginSelector(false)}>
                <div className={"flex flex-col gap-5"}>
                    {paginatedPluginData.loading ? (
                        <div className={"flex items-center justify-center py-10"}>
                            <div className={"w-10 h-10 border-4 border-t-white border-stone-800 rounded-full animate-spin"}/>
                        </div>
                    ) : (
                        <div>
                            {paginatedPluginData.data.length === 0 ? (
                                <div className={"flex flex-col items-center justify-center py-10"}>
                                    <p className={"text-lg"}>You didn't create any plugins yet</p>
                                    <a href={"/plugins/new"} className={"text-lg underline"}>Start with creating one</a>
                                </div>
                            ) : (
                                <Scrollbar style={{ height: 500 }}>
                                    <div className={"grid grid-cols-1 gap-2 pr-2 mt-5"}>
                                        {paginatedPluginData.data.map((plugin, idx) => (
                                            <motion.div
                                                key={plugin.id}
                                                initial={{ opacity: 0, y: 10 }}
                                                animate={{ opacity: 1, y: 0 }}
                                                transition={{ duration: 0.2, delay: idx * 0.1 }}
                                                onClick={() => setSelectedPlugin(plugin)}
                                                className={`px-4 col-span-1 flex flex-row items-center p-1 rounded-2xl ${
                                                    selectedPlugin?.id === plugin.id ? "bg-amber-500" : "border border-amber-500 hover:bg-amber-500 transition-colors duration-200"
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
                            {paginatedPluginData.nextPageToken && (
                                <div className={"flex items-center justify-center pt-3"}>
                                    <motion.button
                                        whileHover={{ scale: 1.05 }}
                                        whileTap={{ scale: 0.95 }}
                                        onClick={() => paginatedPluginData.nextPage(paginatedPluginData.nextPageToken, {append: true, force: true})}
                                        className={"text-sm underline"}
                                    >
                                        Load more
                                    </motion.button>
                                </div>
                            )}
                            <div className={"flex flex-row items-center justify-end pt-5"}>
                                <div className={"flex flex-row items-center justify-between gap-5"}>
                                    <motion.button
                                        whileHover={{ scale: 1.05 }}
                                        whileTap={{ scale: 0.95 }}
                                        onClick={() => setShowPluginSelector(false)}
                                        className={"px-3 py-2 rounded-full bg-white text-black hover:bg-white/80 transition-colors duration-200 disabled:opacity-70 disabled:cursor-not-allowed"}
                                    >
                                        Cancel
                                    </motion.button>
                                    <motion.button
                                        disabled={!selectedPlugin}
                                        whileHover={{ scale: 1.05 }}
                                        whileTap={{ scale: 0.95 }}
                                        onClick={() => setShowPluginSelector(false)}
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
                <div className={"w-full lg:w-1/3"}>
                    <div className={"border-box w-full bg-stone-800 rounded-4xl p-5"}>
                        {selectedPlugin ? (
                            <div className={"flex flex-col gap-5"}>
                                <div className={"flex flex-row items-center justify-between"}>
                                    <p className={"text-lg font-semibold"}>Selected Plugin</p>
                                    <div className={"flex items-center justify-center"}>
                                        <motion.button
                                            whileHover={{ scale: 1.05 }}
                                            whileTap={{ scale: 0.95 }}
                                            onClick={async () => {
                                                await paginatedPluginData.refetch();
                                                setShowPluginSelector(true);
                                            }}
                                            className={"p-2 rounded-full bg-white text-black hover:bg-white/80 transition-colors duration-200 disabled:opacity-70 disabled:cursor-not-allowed"}
                                        >
                                            <Pencil size={15}/>
                                        </motion.button>
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
                                        <p className={"text-md font-semibold"}>{selectedPlugin.name}</p>
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
                                        <p className={"text-md font-semibold"}>{selectedPlugin.filename}</p>
                                    </div>
                                </div>
                            </div>
                        ) : (
                            <div className={"flex flex-col gap-5"}>
                                <div className={"flex flex-row items-center justify-between"}>
                                    <p className={"text-lg font-semibold"}>Selected Plugin</p>
                                    <div className={"flex items-center justify-center"}>
                                        <motion.button
                                            whileHover={{ scale: 1.05 }}
                                            whileTap={{ scale: 0.95 }}
                                            onClick={async () => {
                                                await paginatedPluginData.refetch();
                                                setShowPluginSelector(true);
                                            }}
                                            className={"p-2 rounded-full bg-white text-black hover:bg-white/80 transition-colors duration-200 disabled:opacity-70 disabled:cursor-not-allowed"}
                                        >
                                            <Plus size={15}/>
                                        </motion.button>
                                    </div>
                                </div>
                                <div className={"flex flex-col gap-2"}>
                                    <p className={"text-md font-semibold"}>Plugin not selected</p>
                                    <p className={"text-sm text-gray-400"}>Please select a plugin to associate with this route.</p>
                                </div>
                            </div>
                        )}
                    </div>
                </div>
                <div className={"w-full lg:w-1/3"}>
                    <div className={"border-box w-full bg-stone-800 rounded-4xl p-5"}>
                        {selectedRoute ? (
                            <div className={"flex flex-col gap-5"}>
                                <div className={"flex flex-row items-center justify-between"}>
                                    <p className={"text-lg font-semibold"}>Selected Route</p>
                                    <div className={"flex items-center justify-center"}>
                                        <motion.button
                                            whileHover={{ scale: 1.05 }}
                                            whileTap={{ scale: 0.95 }}
                                            onClick={async () => {
                                                await paginatedRouteData.refetch();
                                                setShowRouteSelector(true);
                                            }}
                                            className={"p-2 rounded-full bg-white text-black hover:bg-white/80 transition-colors duration-200 disabled:opacity-70 disabled:cursor-not-allowed"}
                                        >
                                            <Pencil size={15}/>
                                        </motion.button>
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
                                        <p className={"text-md font-semibold"}>{selectedRoute.path}</p>
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
                                        <p className={"text-md font-semibold"}>{selectedRoute.target_url}</p>
                                    </div>
                                </div>
                            </div>
                        ) : (
                            <div className={"flex flex-col gap-5"}>
                                <div className={"flex flex-row items-center justify-between"}>
                                    <p className={"text-lg font-semibold"}>Selected Route</p>
                                    <div className={"flex items-center justify-center"}>
                                        <motion.button
                                            whileHover={{ scale: 1.05 }}
                                            whileTap={{ scale: 0.95 }}
                                            onClick={async () => {
                                                await paginatedRouteData.refetch();
                                                setShowRouteSelector(true);
                                            }}
                                            className={"p-2 rounded-full bg-white text-black hover:bg-white/80 transition-colors duration-200 disabled:opacity-70 disabled:cursor-not-allowed"}
                                        >
                                            <Plus size={15}/>
                                        </motion.button>
                                    </div>
                                </div>
                                <div className={"flex flex-col gap-2"}>
                                    <p className={"text-md font-semibold"}>Route not selected</p>
                                    <p className={"text-sm text-gray-400"}>Please select a route to associate with this plugin.</p>
                                </div>
                            </div>
                        )}
                    </div>
                </div>
                <div className={"w-full lg:w-1/3"}>
                    <div className={"border-box w-full bg-stone-800 rounded-4xl p-5"}>
                        <div className={"flex flex-col gap-5"}>
                            <div className={"flex flex-row"}>
                                <p className={"text-lg font-semibold"}>Route Plugin Details</p>
                            </div>
                            <div className={"flex flex-col gap-2"}>
                                <p className={"text-md font-semibold"}>Execution Order</p>
                                <Input
                                    required
                                    type={"number"}
                                    value={routePluginFormState.execution_order}
                                    onChange={(e) => setRoutePluginFormState(prev => ({ ...prev, execution_order: parseInt(e.target.value)}))}
                                    className={"w-full px-3 py-2 rounded-lg bg-stone-700 text-white focus:outline-none focus:ring-2 focus:ring-amber-500"}
                                />
                            </div>
                            <div className={"flex flex-col gap-2"}>
                                <p className={"text-md font-semibold"}>Custom JSON Config</p>
                                <p className={"text-sm"}>You will be able to retrieve this config in the WASM plugin through host functions</p>
                                <JsonEditor
                                    data={jsonConfig}
                                    setData={(data) => {
                                        setJsonConfig(data);
                                    }}
                                    theme={githubDarkTheme}
                                />
                            </div>
                            <div className={"flex flex-row items-center justify-end gap-5"}>
                                <motion.button
                                    whileHover={{ scale: 1.05 }}
                                    whileTap={{ scale: 0.95 }}
                                    disabled={!selectedPlugin || !selectedRoute || mutation.loading}
                                    onClick={handleSubmit}
                                    className={"px-3 py-2 rounded-4xl bg-white text-black hover:bg-white/80 transition-colors duration-200 disabled:opacity-70 disabled:cursor-not-allowed"}
                                >
                                    {mutation.loading ? (
                                        <div className={"flex items-center justify-center px-2"}>
                                            <div className={"w-5 h-5 border-2 border-t-black border-white rounded-full animate-spin"}/>
                                        </div>
                                    ) : (
                                        <p className={"text-sm font-semibold"}>Submit</p>
                                    )}
                                </motion.button>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </PageLayout>
    );
}