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

import {
    Field,
    Label,
    Input,
} from "@headlessui/react";
import {motion} from "motion/react";
import NavBar from "@/components/navigation/NavBar";
import React, {useCallback, useEffect, useState} from "react";
import {useRouter} from "next/navigation";
import {useMutation} from "@/hooks/useMutation";
import PageLayout from "@/components/layout/PageLayout";
import {ArrowLeft, Undo2} from "lucide-react";
import Scrollbar from "react-scrollbars-custom";
import {ModalDialog} from "@/components/dialog/ModalDialog";

const initialRouteFormState: Omit<WasmForge.Route, "id" | "created_at" | "enabled"> = {
    path: "/api/example",
    target_url: "http://localhost:9092/api/example",
    idle_conn_timeout: 10,
    tls_handshake_timeout: 15,
    expect_continue_timeout: 5,
};

function NewRoutePageContent() {
    useEffect(() => {
        document.title = "Create new Route - WasmForge";
    });

    const [routeFormData, setRouteFormData] = useState(initialRouteFormState);
    const [createdRoutePath, setCreatedRoutePath] = useState<string | null>(null);
    const [success, setSuccess] = useState(false);
    const mutation = useMutation();
    const router = useRouter();

    const handleSubmit = useCallback(
        async () => {
            const res = await mutation.mutate("http://localhost:8080/api/routes", "POST", JSON.stringify(routeFormData));
            if (res.success) {
                if (res.response) {
                    try{
                        const createdRoute: WasmForge.Route = await res.response.json();
                        setCreatedRoutePath(createdRoute.path);
                    } catch (error) {
                        console.error("Failed to parse response:", error);
                        setCreatedRoutePath(null);
                    }
                }
                setSuccess(true);
            }
        }, [routeFormData, mutation, router]
    );

    const handleCancel = () => {
        setRouteFormData(initialRouteFormState);
    };

    return (
        <div className={"flex flex-col gap-5 w-full mt-20"}>
            <div className={"flex flex-row px-4 items-center justify-between bg-stone-800 rounded-4xl p-3 w-1/3"}>
                <p className={"text-xl font-semibold"}>Creating a new Route</p>
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
                        onClick={handleCancel}
                        className={"px-2 py-2 rounded-full bg-amber-500 text-white hover:bg-amber-500/80 transition-colors duration-200"}
                    >
                        <Undo2 size={15}/>
                    </motion.button>
                </div>
            </div>
            <ModalDialog title={"Error"} visible={!!mutation.error} onClose={() => mutation.setError(null)} >
                <div className={"flex flex-col gap-5"}>
                    <p className={"text-md font-semibold"}>{mutation.error?.message}</p>
                    <Scrollbar style={{ height: 200 }}>
                        <p className={"text-md"}>{mutation.error?.details}</p>
                    </Scrollbar>
                    <div className={"flex flex-row items-end justify-end"}>
                        <motion.button
                            whileHover={{ scale: 1.05 }}
                            whileTap={{ scale: 0.95 }}
                            onClick={() => mutation.setError(null)}
                            className={"px-3 py-2 rounded-full bg-white text-black hover:bg-white/80 transition-colors duration-200"}
                        >
                            Close
                        </motion.button>
                    </div>
                </div>
            </ModalDialog>
            <ModalDialog title={"Route successfully created"} visible={success} onClose={() => router.push(createdRoutePath ? `/routes/route?path=${createdRoutePath}` : "/routes")} >
                <div className={"flex flex-col gap-5"}>
                    <p className={"text-md font-semibold"}>You successfully created a new route!</p>
                    <p className={"text-sm"}>Right now it's disabled. You can navigate to route page to enable it and add plugins</p>
                    <div className={"flex flex-row items-end justify-end"}>
                        <motion.button
                            whileHover={{ scale: 1.05 }}
                            whileTap={{ scale: 0.95 }}
                            onClick={() => router.push(createdRoutePath ? `/routes/route?path=${createdRoutePath}` : "/routes")}
                            className={"px-3 py-2 rounded-full bg-white text-black hover:bg-white/80 transition-colors duration-200"}
                        >
                            Close
                        </motion.button>
                    </div>
                </div>
            </ModalDialog>
            <div className={"flex flex-col lg:flex-row gap-5 w-full"}>
                <div className={"w-full lg:w-2/3"}>
                    <div className={"border-box w-full bg-stone-800 rounded-4xl p-5"}>
                        <div className={"flex flex-col gap-5"}>
                            <div className={"flex flex-row w-full gap-5"}>
                                <Field className={"flex flex-col w-full gap-2"}>
                                    <Label className={"text-md font-semibold"}>Path *</Label>
                                    <Input
                                        required
                                        type={"text"}
                                        value={routeFormData.path}
                                        onChange={(e) => setRouteFormData(prev => ({ ...prev, path: e.target.value }))}
                                        className={"text-md font-semibold px-3 py-2 w-full border border-white focus:border-amber-500 focus:ring focus:ring-amber-500 rounded-xl"}
                                    />
                                </Field>
                                <Field className={"flex flex-col w-full gap-2"}>
                                    <Label className={"text-md font-semibold"}>Target URL *</Label>
                                    <Input
                                        required
                                        type={"text"}
                                        value={routeFormData.target_url}
                                        onChange={(e) => setRouteFormData(prev => ({ ...prev, target_url: e.target.value }))}
                                        className={"text-md font-semibold px-3 py-2 w-full border border-white focus:border-amber-500 focus:ring focus:ring-amber-500 rounded-xl"}
                                    />
                                </Field>
                            </div>
                            <div className={"flex flex-row w-full gap-5"}>
                                <Field className={"flex flex-col w-full gap-2"}>
                                    <Label className={"text-md font-semibold"}>Idle connection timeout *</Label>
                                    <Input
                                        required
                                        value={routeFormData.idle_conn_timeout}
                                        onChange={(e) => setRouteFormData(prev => ({ ...prev, idle_conn_timeout: parseInt(e.target.value) }))}
                                        type={"number"}
                                        className={"text-md font-semibold px-3 py-2 w-full border border-white focus:border-amber-500 focus:ring focus:ring-amber-500 rounded-xl"}
                                    />
                                </Field>
                                <Field className={"flex flex-col w-full gap-2"}>
                                    <Label className={"text-md font-semibold"}>TLS handshake timeout *</Label>
                                    <Input
                                        required
                                        type={"number"}
                                        value={routeFormData.tls_handshake_timeout}
                                        onChange={(e) => setRouteFormData(prev => ({ ...prev, tls_handshake_timeout: parseInt(e.target.value) }))}
                                        className={"text-md font-semibold px-3 py-2 w-full border border-white focus:border-amber-500 focus:ring focus:ring-amber-500 rounded-xl"}
                                    />
                                </Field>
                            </div>
                            <div className={"flex flex-row w-full gap-5"}>
                                <Field className={"flex flex-col w-full gap-2"}>
                                    <Label className={"text-md font-semibold"}>Expect continue timeout *</Label>
                                    <Input
                                        required
                                        type={"number"}
                                        value={routeFormData.expect_continue_timeout}
                                        onChange={(e) => setRouteFormData(prev => ({ ...prev, expect_continue_timeout: parseInt(e.target.value) }))}
                                        className={"text-md font-semibold px-3 py-2 w-full border border-white focus:border-amber-500 focus:ring focus:ring-amber-500 rounded-xl"}
                                    />
                                </Field>
                                <Field className={"flex flex-col w-full gap-2"}>
                                    <Label className={"text-md font-semibold"}>Max idle connections</Label>
                                    <Input
                                        type={"number"}
                                        value={routeFormData.max_idle_cons}
                                        onChange={(e) => setRouteFormData(prev => ({ ...prev, max_idle_cons: parseInt(e.target.value) }))}
                                        className={"text-md font-semibold px-3 py-2 w-full border border-white focus:border-amber-500 focus:ring focus:ring-amber-500 rounded-xl"}
                                    />
                                </Field>
                            </div>
                            <div className={"flex flex-row w-full gap-5"}>
                                <Field className={"flex flex-col w-full gap-2"}>
                                    <Label className={"text-md font-semibold"}>Max idle connections per host</Label>
                                    <Input
                                        type={"number"}
                                        value={routeFormData.max_idle_cons_per_host}
                                        onChange={(e) => setRouteFormData(prev => ({ ...prev, max_idle_cons_per_host: parseInt(e.target.value) }))}
                                        className={"text-md font-semibold px-3 py-2 w-full border border-white focus:border-amber-500 focus:ring focus:ring-amber-500 rounded-xl"}
                                    />
                                </Field>
                                <Field className={"flex flex-col w-full gap-2"}>
                                    <Label className={"text-md font-semibold"}>Max connections per host</Label>
                                    <Input
                                        type={"number"}
                                        value={routeFormData.max_cons_per_host}
                                        onChange={(e) => setRouteFormData(prev => ({ ...prev, max_cons_per_host: parseInt(e.target.value) }))}
                                        className={"text-md font-semibold px-3 py-2 w-full border border-white focus:border-amber-500 focus:ring focus:ring-amber-500 rounded-xl"}
                                    />
                                </Field>
                            </div>
                            <div className={"flex flex-row w-full gap-5"}>
                                <Field className={"flex flex-col gap-2 w-1/2"}>
                                    <Label className={"text-md font-semibold"}>Response header timeout</Label>
                                    <Input
                                        type={"number"}
                                        value={routeFormData.response_header_timeout}
                                        onChange={(e) => setRouteFormData(prev => ({ ...prev, response_header_timeout: parseInt(e.target.value) }))}
                                        className={"text-md font-semibold px-3 py-2 w-full border border-white focus:border-amber-500 focus:ring focus:ring-amber-500 rounded-xl"}
                                    />
                                </Field>
                            </div>
                            <div className={"flex flex-row w-full gap-2 justify-between items-center"}>
                                <div className={"flex"}>
                                    <p className={"text-sm"}>Required fields are marked with *</p>
                                </div>
                                <div className={"flex flex-row gap-2"}>
                                    <motion.button
                                        whileHover={{ scale: 1.05 }}
                                        whileTap={{ scale: 0.95 }}
                                        transition={{ duration: 0.2 }}
                                        onClick={handleCancel}
                                        className={"px-3 py-2 rounded-full bg-white text-black hover:bg-white/80 transition-colors duration-200 disabled:opacity-70 disabled:cursor-not-allowed"}
                                    >
                                        Cancel
                                    </motion.button>
                                    <motion.button
                                        whileHover={{ scale: 1.05 }}
                                        whileTap={{ scale: 0.95 }}
                                        transition={{ duration: 0.2 }}
                                        onClick={handleSubmit}
                                        className={"px-3 py-2 rounded-full bg-white text-black hover:bg-white/80 transition-colors duration-200 disabled:opacity-70 disabled:cursor-not-allowed"}
                                    >
                                        Submit
                                    </motion.button>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
                <div className={"w-full lg:w-1/3"}>
                    <div className={"border-box w-full bg-stone-800 rounded-4xl p-5"}>
                        <div className={"flex flex-col gap-5"}>
                            <div className={"flex flex-row w-full gap-5"}>
                                <p className={"text-lg font-semibold"}>Information</p>
                            </div>
                            <div className={"flex flex-col gap-2"}>
                                <p className={"text-md font-semibold"}>Path</p>
                                <p className={"text-sm"}>Path specifies the prefix of incoming requests. All requests containing this prefix, will be redirected to <strong>Target URL</strong></p>
                            </div>
                            <div className={"flex flex-col gap-2"}>
                                <p className={"text-md font-semibold"}>Target URL</p>
                                <p className={"text-sm"}>Target URL specifies where matching requests will be redirected.</p>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
}

export default function NewRoutePage() {
    const links = [
        { label: "Routes", href: "/routes", active: false },
        { label: "Plugins", href: "/plugins", active: false },
        { label: "Stats", href: "/stats", active: false },
        { label: "Settings", href: "/settings", active: false },
    ];

    return (
        <PageLayout>
            <NavBar links={links} />
            <NewRoutePageContent />
        </PageLayout>
    );
}
