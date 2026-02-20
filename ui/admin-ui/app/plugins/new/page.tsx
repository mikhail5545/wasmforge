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
import PageLayout from "@/components/layout/PageLayout";
import React, {useState, useCallback, useEffect} from "react";
import {useRouter} from "next/navigation";
import {useMutation} from "@/hooks/useMutation";
import {
    Label,
    Field,
    Input,
} from "@headlessui/react";
import {motion} from "motion/react";
import {ArrowLeft, Undo2} from "lucide-react";
import Scrollbar from "react-scrollbars-custom";
import {ModalDialog} from "@/components/dialog/ModalDialog";

const initialPluginFormState: Omit<WasmForge.Plugin, "id" | "created_at" | "checksum"> = {
    name: "my_new_plugin",
    filename: "my_new_plugin.wasm",
};

export default function NewPluginPage() {
    const links = [
        { label: "Routes", href: "/routes", active: false },
        { label: "Plugins", href: "/plugins", active: false },
        { label: "Settings", href: "/settings", active: false },
    ];

    useEffect(() => {
        document.title = "Create new Plugin - WasmForge";
    });

    const [pluginFormData, setPluginFormData] = useState(initialPluginFormState);
    const [file, setFile] = useState<File | null>(null);
    const [success, setSuccess] = useState(false);

    const router = useRouter();
    const mutation = useMutation();

    const [createdPluginName, setCreatedPluginName] = useState<string | null>(null);

    const handleSubmit = useCallback(
        async() => {
            setSuccess(false);
            if (!file) {
                return;
            }

            const formData = new FormData();
            formData.append("wasm_file", file);
            formData.append("metadata", JSON.stringify(pluginFormData));

            const res = await mutation.mutate("http://localhost:8080/api/plugins", "POST", formData);
            if (!res.success && mutation.error) {
                formData.delete("wasm_file");
                formData.delete("metadata");
                return;
            }
            if (res.success && res.response) {
                try{
                    const createdPlugin: WasmForge.Plugin = await res.response.json();
                    setCreatedPluginName(createdPlugin.name);
                }catch (error) {
                    console.log("Error parsing response:", error);
                    setCreatedPluginName(null);
                }
                setSuccess(true);
            }
        }, [file, pluginFormData, mutation]
    );

    const handleCancel = () => {
        setPluginFormData(initialPluginFormState);
        setFile(null);
    };

    return (
        <PageLayout>
            <NavBar links={links}/>
            <div className={"flex flex-col gap-5 w-full mt-20"}>
                <div className={"flex flex-row px-4 items-center justify-between bg-stone-800 rounded-4xl p-3 w-1/3"}>
                    <p className={"text-xl font-semibold"}>Creating a new Plugin</p>
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
                <ModalDialog title={"Plugin successfully created"} visible={success} onClose={() => router.push(createdPluginName ? `/plugins/plugin?name=${createdPluginName}` : "/plugins")} >
                    <div className={"flex flex-col gap-5"}>
                        <p className={"text-md font-semibold"}>You successfully created a new plugin!</p>
                        <p className={"text-sm"}>You can now attach it to a route</p>
                        <div className={"flex flex-row items-end justify-end"}>
                            <motion.button
                                whileHover={{ scale: 1.05 }}
                                whileTap={{ scale: 0.95 }}
                                onClick={() => router.push(createdPluginName ? `/plugins/plugin?name=${createdPluginName}` : "/plugins")}
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
                                        <Label className={"text-md font-semibold"}>Name *</Label>
                                        <Input
                                            required
                                            type={"text"}
                                            value={pluginFormData.name}
                                            onChange={(e) => setPluginFormData(prev => ({ ...prev, name: e.target.value }))}
                                            className={"text-md font-semibold px-3 py-2 w-full border border-white focus:border-amber-500 focus:ring focus:ring-amber-500 rounded-xl"}
                                        />
                                    </Field>
                                    <Field className={"flex flex-col w-full gap-2"}>
                                        <Label className={"text-md font-semibold"}>Filename *</Label>
                                        <Input
                                            required
                                            type={"text"}
                                            value={pluginFormData.filename}
                                            onChange={(e) => setPluginFormData(prev => ({ ...prev, filename: e.target.value }))}
                                            className={"text-md font-semibold px-3 py-2 w-full border border-white focus:border-amber-500 focus:ring focus:ring-amber-500 rounded-xl"}
                                        />
                                    </Field>
                                </div>
                                <div className={"flex flex-row w-full gap-5"}>
                                    <Field className={"flex flex-col w-full gap-2 lg:w-1/2"}>
                                        <Label className={"text-md font-semibold"}>WASM File *</Label>
                                        <Input
                                            required
                                            type={"file"}
                                            accept={".wasm"}
                                            onChange={(e) => setFile(e.target.files ? e.target.files[0] : null)}
                                            className={"text-md font-semibold px-3 py-5 w-full border border-white focus:border-amber-500 focus:ring focus:ring-amber-500 rounded-xl"}
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
                                    <p className={"text-md font-semibold"}>Name and Filename</p>
                                    <p className={"text-sm"}>Name and Filename are just unique identifiers for plugin that will help you to track and manage them.</p>
                                </div>
                                <div className={"flex flex-col gap-2"}>
                                    <p className={"text-md font-semibold"}>WASM File</p>
                                    <p className={"text-sm"}>WASM File will be saved and used when you will decide to attach this plugin to some route. In this case, it will be uploaded into WASM runtime and registered as a new middleware for selected route.</p>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </PageLayout>
    );
}