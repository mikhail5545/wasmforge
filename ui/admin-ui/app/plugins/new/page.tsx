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
import React, {useState, SubmitEvent, useCallback} from "react";
import {useRouter} from "next/navigation";
import {ErrorDialog} from "@/components/dialog/ErrorDialog";
import {
    Fieldset,
    Label,
    Legend,
    Field,
    Input,
} from "@headlessui/react";
import {motion} from "motion/react";
import {InfoDialog} from "@/components/dialog/InfoDialog";

const initialPluginFormState: Omit<WasmForge.Plugin, "id" | "created_at" | "checksum"> = {
    name: "",
    filename: "",
};

export default function NewPluginPage() {
    const links = [
        { label: "Dashboard", href: "/dashboard" },
        { label: "Users", href: "/users" },
        { label: "Settings", href: "/settings" },
    ];

    const [pluginFormData, setPluginFormData] = useState(initialPluginFormState);
    const [file, setFile] = useState<File | null>(null);
    const [success, setSuccess] = useState(false);

    const router = useRouter();

    const onSubmit = async(e: SubmitEvent<HTMLFormElement>) => {
        e.preventDefault();

        const form = new FormData();
        if (file) {
            form.append("wasm_file", file);
        }

        form.append("metadata", JSON.stringify(pluginFormData));

        const response = await fetch("http://localhost:8080/api/plugins", {
            method: "POST",
            body: form,
        });
        if (!response.ok) {
            const errorData = await response.json();
            throw new Error(errorData.error.details || "Failed to create plugin");
        }
        if (response.ok) {
            setSuccess(true);
        }

    };

    const handleCancel = useCallback(() => {
        setFile(null);
        setPluginFormData(initialPluginFormState);
        router.push("/plugins");
    }, [router, setFile, setPluginFormData]);

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
                {/*<ErrorDialog title={mutation.error?.name || "An error occurred"} message={mutation.error?.message || ""} isOpen={!!mutation.error} onClose={() => mutation.setError(null)} />*/}
                <InfoDialog title={"Success"} message={"Plugin successfully created"} isOpen={success} onClose={() => router.push("/plugins")}/>
                <div className={"px-5 md:px-15 lg:px-30 py-10"}>
                    <div className={"px-20"}>
                        <form onSubmit={onSubmit}>
                            <Fieldset className={"space-y-6 rounded-xl bg-white/5 p-6 sm:p-8"}>
                                <Legend className={"text-lg font-semibold text-white"}>
                                    Plugin details
                                    <p className={"text-sm font-semibold text-white/50"}>Required fields are marked with *</p>
                                </Legend>
                                <Field>
                                    <Label className={"text-base/7 font-semibold text-white"}>Name*</Label>
                                    <Input
                                        required
                                        type={"text"}
                                        value={pluginFormData.name}
                                        onChange={(e) => setPluginFormData(prev => ({ ...prev, name: e.target.value }))}
                                        className={"mt-1 w-full rounded-md border border-stone-700 bg-stone-800/50 px-3 py-2 text-sm text-white focus:border-blue-500 focus:ring focus:ring-blue-500/50"}
                                    />
                                </Field>
                                <Field>
                                    <Label className={"text-base/7 font-semibold text-white"}>Filename*</Label>
                                    <Input
                                        required
                                        type={"text"}
                                        value={pluginFormData.filename}
                                        onChange={(e) => setPluginFormData(prev => ({ ...prev, filename: e.target.value }))}
                                        className={"mt-1 w-full rounded-md border border-stone-700 bg-stone-800/50 px-3 py-2 text-sm text-white focus:border-blue-500 focus:ring focus:ring-blue-500/50"}
                                    />
                                </Field>
                                <Field>
                                    <Label className={"text-base/7 font-semibold text-white"}>Upload file</Label>
                                    <Input
                                        required
                                        type={"file"}
                                        accept={".wasm"}
                                        onChange={(e) => {
                                            const selectedFile = e.target.files?.[0] || null;
                                            setFile(selectedFile);
                                        }}
                                        className={"mt-1 w-full rounded-md border border-stone-700 bg-stone-800/50 px-3 py-5 text-sm text-white focus:border-blue-500 focus:ring focus:ring-blue-500/50"}
                                    />
                                </Field>
                                <div className={"flex flex-row gap-5 pt-4"}>
                                    <motion.button
                                        whileHover={{ scale: 1.05 }}
                                        whileTap={{ scale: 0.95 }}
                                        transition={{ duration: 0.2 }}
                                        type={"submit"}
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