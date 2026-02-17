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
    Fieldset,
    Field,
    Description,
    Legend,
    Label,
    Input,
} from "@headlessui/react";
import {motion} from "motion/react";
import NavBar from "@/components/navigation/NavBar";
import React, {useCallback, useState, ChangeEvent} from "react";
import {useRouter} from "next/navigation";
import {useMutation} from "@/hooks/useMutation";
import {ErrorDialog} from "@/components/dialog/ErrorDialog";

const initialRouteFormState: Omit<WasmForge.Route, "id" | "created_at" | "enabled"> = {
    path: "/api/example",
    target_url: "http://localhost:8080/api/example",
    idle_conn_timeout: 10,
    tls_handshake_timeout: 15,
    expect_continue_timeout: 5,
};

export default function NewRoutePage() {
    const links = [
        { label: "Dashboard", href: "/dashboard" },
        { label: "Routes", href: "/routes" },
        { label: "Plugins", href: "/plugins" },
    ];

    const [routeFormData, setRouteFormData] = useState(initialRouteFormState);
    const mutation = useMutation();
    const router = useRouter();

    const handleSubmit = useCallback(
        async () => {
            const { success: submissionSuccess } = await mutation.mutate("http://localhost:8080/api/routes", "POST", JSON.stringify(routeFormData));
            if (submissionSuccess) {
                setRouteFormData(initialRouteFormState);
                router.push("/routes");
            }
        }, [routeFormData, mutation, router]
    );

    const handleCancel = useCallback(
        () => {
            setRouteFormData(initialRouteFormState);
            router.push("/routes");
        }, [router]
    );

    return (
        <div className={"flex min-h-screen bg-stone-950 font-mono text-white"}>
            <div className={"flex flex-col w-full"}>
                <NavBar
                    title={"WasmForge"}
                    links={links}
                />
                <div className={"px-5 md:px-15 lg:px-30 py-10"}>
                    <ErrorDialog
                        title={mutation.error ? "Error fetching routes" : ""}
                        message={mutation.error ? mutation.error.message : ""}
                        isOpen={!!mutation.error}
                        onClose={() => mutation.setError(null)}
                    />
                    <div className={"px-20"}>
                        <form onSubmit={e => e.preventDefault()}>
                            <Fieldset className={"space-y-6 rounded-xl bg-white/5 p-6 sm:p-8"}>
                                <Legend className={"text-lg font-semibold text-white"}>
                                    Route details
                                    <p className={"text-sm font-semibold text-white/50"}>Required fields are marked with *</p>
                                </Legend>
                                <Field>
                                    <Label className={"text-base/7 font-semibold text-white"}>Path*</Label>
                                    <Description className={"text-sm text-white/50"}>The path which will be used to determine where to redirect incoming request</Description>
                                    <Input
                                        type={"text"}
                                        name={"path"}
                                        className={"mt-1 w-full rounded-md border border-stone-700 bg-stone-800/50 px-3 py-2 text-sm text-white focus:border-blue-500 focus:ring focus:ring-blue-500/50"}
                                        value={routeFormData.path}
                                        onChange={(e: ChangeEvent<HTMLInputElement>) => setRouteFormData(prev => ({ ...prev, path: e.target.value }))}
                                    />
                                </Field>
                                <Field>
                                    <Label className={"text-base/7 font-semibold text-white"}>Target URL*</Label>
                                    <Description className={"text-sm text-white/50"}>Where to redirect requests</Description>
                                    <Input
                                        type={"text"}
                                        name={"path"}
                                        className={"mt-1 w-full rounded-md border border-stone-700 bg-stone-800/50 px-3 py-2 text-sm text-white focus:border-blue-500 focus:ring focus:ring-blue-500/50"}
                                        value={routeFormData.target_url}
                                        onChange={(e: ChangeEvent<HTMLInputElement>) => setRouteFormData(prev => ({ ...prev, target_url: e.target.value }))}
                                    />
                                </Field>
                                <Field>
                                    <Label className={"text-base/7 font-semibold text-white"}>Idle connection timeout*</Label>
                                    <Input
                                        type={"number"}
                                        name={"path"}
                                        className={"mt-1 w-full rounded-md border border-stone-700 bg-stone-800/50 px-3 py-2 text-sm text-white focus:border-blue-500 focus:ring focus:ring-blue-500/50"}
                                        value={routeFormData.idle_conn_timeout}
                                        onChange={(e: ChangeEvent<HTMLInputElement>) => setRouteFormData(prev => ({ ...prev, idle_conn_timeout: parseInt(e.target.value) }))}
                                    />
                                </Field>
                                <Field>
                                    <Label className={"text-base/7 font-semibold text-white"}>TLS handshake timeout*</Label>
                                    <Description className={"text-sm text-white/50"}>TLS handshake timeout for this specific route in seconds</Description>
                                    <Input
                                        type={"number"}
                                        name={"path"}
                                        className={"mt-1 w-full rounded-md border border-stone-700 bg-stone-800/50 px-3 py-2 text-sm text-white focus:border-blue-500 focus:ring focus:ring-blue-500/50"}
                                        value={routeFormData.tls_handshake_timeout}
                                        onChange={(e: ChangeEvent<HTMLInputElement>) => setRouteFormData(prev => ({ ...prev, tls_handshake_timeout: parseInt(e.target.value) }))}
                                    />
                                </Field>
                                <Field>
                                    <Label className={"text-base/7 font-semibold text-white"}>Expected continue timeout*</Label>
                                    <Description className={"text-sm text-white/50"}>Expected continue timeout for this specific route in seconds</Description>
                                    <Input
                                        type={"number"}
                                        name={"path"}
                                        className={"mt-1 w-full rounded-md border border-stone-700 bg-stone-800/50 px-3 py-2 text-sm text-white focus:border-blue-500 focus:ring focus:ring-blue-500/50"}
                                        placeholder={"5"}
                                        value={routeFormData.expect_continue_timeout}
                                        onChange={(e: ChangeEvent<HTMLInputElement>) => setRouteFormData(prev => ({ ...prev, expect_continue_timeout: parseInt(e.target.value) }))}
                                    />
                                </Field>
                                <Field>
                                    <Label className={"text-base/7 font-semibold text-white"}>Max idle connections</Label>
                                    <Input
                                        type={"number"}
                                        name={"path"}
                                        value={routeFormData.max_idle_cons}
                                        placeholder={"100"}
                                        className={"mt-1 w-full rounded-md border border-stone-700 bg-stone-800/50 px-3 py-2 text-sm text-white focus:border-blue-500 focus:ring focus:ring-blue-500/50"}
                                        onChange={(e: ChangeEvent<HTMLInputElement>) => setRouteFormData(prev => ({ ...prev, max_idle_cons: parseInt(e.target.value) }))}
                                    />
                                </Field>
                                <Field>
                                    <Label className={"text-base/7 font-semibold text-white"}>Max idle connections per host</Label>
                                    <Input
                                        type={"number"}
                                        name={"path"}
                                        className={"mt-1 w-full rounded-md border border-stone-700 bg-stone-800/50 px-3 py-2 text-sm text-white focus:border-blue-500 focus:ring focus:ring-blue-500/50"}
                                        placeholder={"100"}
                                        value={routeFormData.max_idle_cons_per_host}
                                        onChange={(e: ChangeEvent<HTMLInputElement>) => setRouteFormData(prev => ({ ...prev, max_idle_cons_per_host: parseInt(e.target.value) }))}
                                    />
                                </Field>
                                <Field>
                                    <Label className={"text-base/7 font-semibold text-white"}>Max connections per host</Label>
                                    <Input
                                        type={"number"}
                                        name={"path"}
                                        className={"mt-1 w-full rounded-md border border-stone-700 bg-stone-800/50 px-3 py-2 text-sm text-white focus:border-blue-500 focus:ring focus:ring-blue-500/50"}
                                        placeholder={"100"}
                                        value={routeFormData.max_cons_per_host}
                                        onChange={(e: ChangeEvent<HTMLInputElement>) => setRouteFormData(prev => ({ ...prev, max_cons_per_host: parseInt(e.target.value) }))}
                                    />
                                </Field>
                                <Field>
                                    <Label className={"text-base/7 font-semibold text-white"}>Response header timeout</Label>
                                    <Description className={"text-sm text-white/50"}>Response header timeout for this specific route in seconds</Description>
                                    <Input
                                        type={"number"}
                                        name={"path"}
                                        className={"mt-1 w-full rounded-md border border-stone-700 bg-stone-800/50 px-3 py-2 text-sm text-white focus:border-blue-500 focus:ring focus:ring-blue-500/50"}
                                        placeholder={"5"}
                                        value={routeFormData.response_header_timeout}
                                        onChange={(e: ChangeEvent<HTMLInputElement>) => setRouteFormData(prev => ({ ...prev, response_header_timeout: parseInt(e.target.value) }))}
                                    />
                                </Field>
                                <div className={"flex flex-row gap-5 pt-4"}>
                                    <motion.button
                                        whileHover={{ scale: 1.05 }}
                                        whileTap={{ scale: 0.95 }}
                                        transition={{ duration: 0.2 }}
                                        type={"submit"}
                                        onClick={handleSubmit}
                                        className={"w-1/2 px-4 py-2 bg-white text-black rounded-4xl text-sm hover:bg-stone-800 hover:text-white border border-white transition-colors duration-200"}
                                    >
                                        Submit
                                    </motion.button>
                                    <motion.button
                                        whileHover={{ scale: 1.05 }}
                                        whileTap={{ scale: 0.95 }}
                                        transition={{ duration: 0.2 }}
                                        type={"button"}
                                        onClick={handleCancel}
                                        className={"w-1/2 px-4 py-2 bg-stone-800 text-white rounded-4xl text-sm hover:bg-stone-700 transition-colors duration-200"}
                                    >
                                        Cancel
                                    </motion.button>
                                </div>
                            </Fieldset>
                        </form>
                    </div>
                </div>
            </div>
        </div>
    );
}