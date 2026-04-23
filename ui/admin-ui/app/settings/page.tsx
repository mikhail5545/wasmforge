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

import {useCallback, useState} from "react";
import PageLayout from "@/components/layout/PageLayout";
import NavBar from "@/components/navigation/NavBar";
import {AnimatePresence, motion} from "motion/react";
import {EthernetPort, Clock, Power, RefreshCw, Pencil, Check, X, ChevronDown, ArrowLeft, Trash} from "lucide-react";
import React from "react";
import {useData} from "@/hooks/useData";
import {Field, Input, Select} from "@headlessui/react";
import {useMutation} from "@/hooks/useMutation";
import {ModalDialog} from "@/components/dialog/ModalDialog";
import Scrollbar from "react-scrollbars-custom";

export default function SettingsPage() {
    const links = [
        { label: "Routes", href: "/routes", active: false },
        { label: "Plugins", href: "/plugins", active: false },
        { label: "Stats", href: "/stats", active: false },
        { label: "Settings", href: "/settings", active: true },
    ];

    const proxyServerStatus = useData<WasmForge.ProxyServerStatus>("http://localhost:8080/api/proxy/config", "status");
    const mutation = useMutation();

    const [editPort, setEditPort] = useState(false);
    const [editReadHeaderTimeout, setEditReadHeaderTimeout] = useState(false);
    const [showSetupTLS, setShowSetupTLS] = useState(false);
    const [proxyServerConfig, setProxyServerConfig] = useState<Omit<
        WasmForge.ProxyServerConfig, "tls_cert_path" | "tls_cert_hash" | "tls_key_path" | "tls_key_hash" | "tls_enabled"
    > | null>(null);
    const [tlsGenerateConfig, setTlsGenerateConfig] = useState<{ common_name: string; valid_days: number; rsa_bits: 2048 | 4096}>({common_name: "localhost", valid_days: 365, rsa_bits: 2048});
    const [certFile, setCertFile] = useState<File | null>(null);
    const [keyFile, setKeyFile] = useState<File | null>(null);
    const [showDeleteTLSConfirmation, setShowDeleteTLSConfirmation] = useState(false);

    const handleSubmitChange = useCallback(
        async () => {
            if (!proxyServerConfig) return;

            const res = await mutation.mutate("http://localhost:8080/api/proxy/config", "PUT", JSON.stringify(proxyServerConfig));
            if (res.success) {
                await proxyServerStatus.refetch();
                setEditPort(false);
                setEditReadHeaderTimeout(false);
            }
        }, [proxyServerConfig, mutation, proxyServerStatus]
    );

    const startProxyServer = useCallback(
        async () => {
            const res = await mutation.mutate("http://localhost:8080/api/proxy/server/start", "POST");
            if (res.success) {
                await proxyServerStatus.refetch();
            }
        }, [mutation, proxyServerStatus]
    );

    const stopProxyServer = useCallback(
        async () => {
            const res = await mutation.mutate("http://localhost:8080/api/proxy/server/stop", "POST");
            if (res.success) {
                await proxyServerStatus.refetch();
            }
        }, [mutation, proxyServerStatus]
    );

    const restartProxyServer = useCallback(
        async () => {
            const res = await mutation.mutate("http://localhost:8080/api/proxy/server/restart", "POST");
            if (res.success) {
                await proxyServerStatus.refetch();
            }
        }, [mutation, proxyServerStatus]
    );

    const generateTLSCertificates = useCallback(
        async () => {
            const res = await mutation.mutate("http://localhost:8080/api/proxy/certs/generate", "POST", JSON.stringify(tlsGenerateConfig));
            if (res.success) {
                await proxyServerStatus.refetch();
                setShowSetupTLS(false);
            }
        }, [mutation, proxyServerStatus, tlsGenerateConfig]
    );

    const uploadTLSCertificates = useCallback(
        async () => {
            if (!certFile || !keyFile) {
                mutation.setError({ code: "", message: "Both certificate and key files are required", details: "" });
                return;
            }
            const formData = new FormData();
            formData.append("cert_file", certFile);
            formData.append("key_file", keyFile);

            const res = await mutation.mutate("http://localhost:8080/api/proxy/certs/upload", "POST", formData);
            if (res.success) {
                await proxyServerStatus.refetch();
                setShowSetupTLS(false);
                setCertFile(null);
                setKeyFile(null);
            }
        }, [certFile, keyFile, mutation, proxyServerStatus]
    );

    const disableTLS = useCallback(
        async () => {
            const res = await mutation.mutate("http://localhost:8080/api/proxy/certs/disable", "POST");
            if (res.success) {
                await proxyServerStatus.refetch();
            }
        }, [mutation, proxyServerStatus]
    );

    const globalError = proxyServerStatus.error || mutation.error;
    const unsetError = useCallback(
        async () => {
        mutation.setError(null);
        if (proxyServerStatus.error) {
            await proxyServerStatus.refetch();
            setShowDeleteTLSConfirmation(false);
        }
    }, [mutation, proxyServerStatus.error]);

    return (
        <PageLayout>
            <NavBar links={links}/>
            <ModalDialog title={"Error"} visible={!!mutation.error} onClose={unsetError} >
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
            <ModalDialog title={"Delete TLS Configuration"} visible={showDeleteTLSConfirmation} onClose={() => setShowDeleteTLSConfirmation(false)} >
                <div className={"flex flex-col gap-5"}>
                    <p className={"text-md"}>This will disable TLS for proxy server and remove all uploaded or generated certificates. You will be able to set up TLS again any time.</p>
                    <div className={"flex flex-row items-end justify-end"}>
                        <motion.button
                            whileHover={{ scale: 1.05 }}
                            whileTap={{ scale: 0.95 }}
                            onClick={disableTLS}
                            className={"px-3 py-2 rounded-full bg-red-500 text-white hover:bg-red-500/80 transition-colors duration-200"}
                        >
                            Delete TLS Configuration
                        </motion.button>
                    </div>
                </div>
            </ModalDialog>
            <div className={"flex flex-col gap-5 w-full mt-20"}>
                <div className={"flex flex-row w-full lg:w-2/3 gap-5"}>
                    <div className={"flex flex-row px-4 items-center justify-between bg-stone-800 rounded-4xl p-3 w-2/3"}>
                        <p className={"text-xl font-semibold"}>Proxy Server Settings</p>
                        {proxyServerStatus.loading ? (
                            <div className={"flex items-center justify-center px-5"}>
                                <div className={"w-5 h-5 border-2 border-t-white border-stone-600 rounded-full animate-spin"}/>
                            </div>
                        ) : (
                            <div className={"flex flex-row gap-2"}>
                                <div className={"flex flex-row items-center gap-2 bg-amber-500 rounded-4xl"}>
                                    <p className={"text-white text-sm pl-4"}>Restart</p>
                                    <motion.button
                                        whileHover={{ scale: 1.05 }}
                                        whileTap={{ scale: 0.95 }}
                                        onClick={restartProxyServer}
                                        disabled={mutation.loading}
                                        className={"px-2 py-2 rounded-full bg-white text-black hover:bg-white/80 transition-colors duration-200"}
                                    >
                                        <RefreshCw size={15}/>
                                    </motion.button>
                                </div>
                                <div className={"flex flex-row items-center gap-2 bg-amber-500 rounded-4xl"}>
                                    <p className={"text-white text-sm pl-4"}>{proxyServerStatus.data.running ? "Stop" : "Start"}</p>
                                    <motion.button
                                        whileHover={{ scale: 1.05 }}
                                        whileTap={{ scale: 0.95 }}
                                        onClick={proxyServerStatus.data.running ? stopProxyServer : startProxyServer}
                                        disabled={mutation.loading}
                                        className={"px-2 py-2 rounded-full bg-white text-black hover:bg-white/80 transition-colors duration-200"}
                                    >
                                        <Power size={15}/>
                                    </motion.button>
                                </div>
                            </div>
                        )}
                    </div>
                    <div className={"flex flex-row px-4 bg-stone-800 rounded-4xl p-3 w-1/3 lg:mr-3"}>
                        {proxyServerStatus.loading ? (
                            <div className={"flex items-center justify-center px-5"}>
                                <div className={"w-5 h-5 border-2 border-t-white border-stone-600 rounded-full animate-spin"}/>
                            </div>
                        ) : (
                            <div className={"flex flex-row items-center justify-center gap-2 w-full"}>
                                <p className={"text-sm font-semibold"}>Proxy Server is <strong>{proxyServerStatus.data.running ? "Running" : "Stopped"}</strong></p>
                            </div>
                        )}
                    </div>
                </div>
                <div className={"flex flex-col lg:flex-row gap-5 w-full"}>
                    <div className={"w-full lg:w-1/3"}>
                        <div className={"border-box w-full bg-stone-800 rounded-4xl p-5"}>
                            {proxyServerStatus.loading ? (
                                <div className={"flex items-center justify-center py-20"}>
                                    <div className={"w-10 h-10 border-4 border-t-white border-stone-800 rounded-full animate-spin"}/>
                                </div>
                            ) : (
                                <div className={"flex flex-col gap-5"}>
                                    <div className={"flex flex-row w-full justify-between"}>
                                        <div className={"flex flex-row gap-5 w-full"}>
                                            <div className={"flex items-center justify-center"}>
                                                <div className={"bg-amber-500 p-2 rounded-full"}>
                                                    <EthernetPort size={15}/>
                                                </div>
                                            </div>
                                            <div className={"flex flex-col gap-1"}>
                                                <p className={"text-md px-2"}>Port</p>
                                                <AnimatePresence mode={"wait"}>
                                                    {editPort ? (
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
                                                                value={proxyServerConfig?.listen_port || 0}
                                                                disabled={!editPort}
                                                                onChange={(e) => setProxyServerConfig(prev => prev ? ({ ...prev, listen_port: parseInt(e.target.value) }) : null)}
                                                                className={`border border-white text-md font-semibold px-2 w-full focus:border-amber-500 focus:ring focus:ring-amber-500 rounded-xl`}
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
                                                            {proxyServerStatus.data.config.listen_port}
                                                        </motion.p>
                                                    )}
                                                </AnimatePresence>
                                            </div>
                                        </div>
                                        <div className={"flex items-center justify-center gap-2"}>
                                            <AnimatePresence mode={"wait"}>
                                                {editPort && (
                                                    <motion.button
                                                        key={"submit-port"}
                                                        whileHover={{ scale: 1.05 }}
                                                        whileTap={{ scale: 0.95 }}
                                                        disabled={mutation.loading}
                                                        onClick={handleSubmitChange}
                                                        className={"p-2 rounded-full bg-white text-black hover:bg-white/80 transition-colors duration-200 disabled:opacity-70 disabled:cursor-not-allowed"}
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
                                                disabled={editReadHeaderTimeout || mutation.loading}
                                                onClick={editPort ? () => {setEditPort(false); setProxyServerConfig(proxyServerStatus.data.config);} : () => { setEditPort(true); setProxyServerConfig(proxyServerStatus.data.config); }}
                                                className={"p-2 rounded-full bg-white text-black hover:bg-white/80 transition-colors duration-200 disabled:opacity-70 disabled:cursor-not-allowed"}
                                            >
                                                {editPort ? <X size={15}/> : <Pencil size={15}/> }
                                            </motion.button>
                                        </div>
                                    </div>
                                    <div className={"flex flex-row w-full justify-between"}>
                                        <div className={"flex flex-row w-full gap-5"}>
                                            <div className={"flex items-center justify-center"}>
                                                <div className={"bg-amber-500 p-2 rounded-full"}>
                                                    <Clock size={15}/>
                                                </div>
                                            </div>
                                            <div className={"flex flex-col"}>
                                                <p className={"text-md px-2"}>Read header timeout</p>
                                                <AnimatePresence mode={"wait"}>
                                                    {editReadHeaderTimeout ? (
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
                                                                value={proxyServerConfig?.read_header_timeout || 0}
                                                                disabled={!editReadHeaderTimeout}
                                                                onChange={(e) => setProxyServerConfig(prev => prev ? ({ ...prev, read_header_timeout: parseInt(e.target.value) }) : null)}
                                                                className={`border border-white text-md font-semibold px-2 w-full focus:border-amber-500 focus:ring focus:ring-amber-500 rounded-xl`}
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
                                                            {proxyServerStatus.data.config.read_header_timeout}
                                                        </motion.p>
                                                    )}
                                                </AnimatePresence>
                                            </div>
                                        </div>
                                        <div className={"flex items-center justify-center gap-2"}>
                                            <AnimatePresence mode={"wait"}>
                                                {editReadHeaderTimeout && (
                                                    <motion.button
                                                        key={"submit-port"}
                                                        whileHover={{ scale: 1.05 }}
                                                        whileTap={{ scale: 0.95 }}
                                                        disabled={mutation.loading}
                                                        onClick={handleSubmitChange}
                                                        className={"p-2 rounded-full bg-white text-black hover:bg-white/80 transition-colors duration-200 disabled:opacity-70 disabled:cursor-not-allowed"}
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
                                                disabled={editPort || mutation.loading}
                                                onClick={editReadHeaderTimeout ? () => {setEditReadHeaderTimeout(false); setProxyServerConfig(proxyServerStatus.data.config);}  : () => {setEditReadHeaderTimeout(true); setProxyServerConfig(proxyServerStatus.data.config);} }
                                                className={"p-2 rounded-full bg-white text-black hover:bg-white/80 transition-colors duration-200 disabled:opacity-70 disabled:cursor-not-allowed"}
                                            >
                                                {editReadHeaderTimeout ? <X size={15}/> : <Pencil size={15}/> }
                                            </motion.button>
                                        </div>
                                    </div>
                                </div>
                            )}
                        </div>
                    </div>
                    <div className={"w-full lg:w-2/3"}>
                        <AnimatePresence mode={"wait"}>
                            {showSetupTLS && (
                                <motion.div
                                    key={"setup-tls-configuration"}
                                    initial={{ opacity: 0, y: -10 }}
                                    animate={{ opacity: 1, y: 0 }}
                                    exit={{ opacity: 0, y: -10 }}
                                    transition={{ duration: 0.2 }}
                                    className={"border-box w-full bg-stone-800 rounded-4xl p-5 mb-5"}
                                >
                                    <div className={"flex flex-col gap-5"}>
                                        <div className={"flex flex-row items-center"}>
                                            <div className={"flex flex-col gap-2"}>
                                                <div className={"flex flex-row items-center gap-4"}>
                                                    <div className={"flex items-center justify-center"}>
                                                        <motion.button
                                                            whileHover={{ scale: 1.05 }}
                                                            whileTap={{ scale: 0.95 }}
                                                            onClick={() => {setShowSetupTLS(false); setTlsGenerateConfig({common_name: "localhost", valid_days: 365, rsa_bits: 2048});}}
                                                            className={"p-2 rounded-full bg-white text-black hover:bg-white/80 transition-colors duration-200"}
                                                        >
                                                            <ArrowLeft size={10} />
                                                        </motion.button>
                                                    </div>
                                                    <p className={"text-xl font-semibold"}>TLS setup</p>
                                                </div>
                                                <p className={"text-sm"}>To setup TLS, you can either upload pre-generated certificate key pair, or generate self-signed certificates for development.</p>
                                            </div>
                                        </div>
                                        <div className={"flex flex-col gap-5"}>
                                            <p className={"text-md font-semibold"}>Upload certificates</p>
                                            <div className={"flex flex-row gap-5"}>
                                                <div className={"flex flex-col w-1/2 gap-2"}>
                                                    <p className={"text-sm"}>Upload <strong>certificate</strong></p>
                                                    <Input
                                                        required
                                                        type={"file"}
                                                        accept={".crt,.pem"}
                                                        onChange={(e) => setCertFile(e.target.files ? e.target.files[0] : null)}
                                                        className={"border border-white px-3 py-3 rounded-2xl"}
                                                    />
                                                </div>
                                                <div className={"flex flex-col w-1/2 gap-2"}>
                                                    <p className={"text-sm"}>Upload <strong>key</strong></p>
                                                    <Input
                                                        required
                                                        type={"file"}
                                                        accept={".key,.pem"}
                                                        onChange={(e) => setKeyFile(e.target.files ? e.target.files[0] : null)}
                                                        className={"border border-white px-3 py-3 rounded-2xl"}
                                                    />
                                                </div>
                                            </div>
                                            <div className={"flex flex-row items-end justify-end"}>
                                                <div className={"flex flex-row"}>
                                                    <motion.button
                                                        whileHover={{ scale: 1.05 }}
                                                        whileTap={{ scale: 0.95 }}
                                                        disabled={mutation.loading}
                                                        onClick={uploadTLSCertificates}
                                                        className={"px-3 py-2 rounded-full bg-white text-black hover:bg-white/80 transition-colors duration-200 disabled:opacity-70 disabled:cursor-not-allowed"}
                                                    >
                                                        Upload
                                                    </motion.button>
                                                </div>
                                            </div>
                                        </div>
                                        <div className={"flex flex-col gap-5 py-2"}>
                                            <p className={"text-md font-semibold"}>Generate self-signed certificates</p>
                                            <div className={"flex flex-row gap-5"}>
                                                <div className={"flex flex-col w-1/2 gap-2"}>
                                                    <p className={"text-sm font-semibold"}>Common Name</p>
                                                    <Input
                                                        required
                                                        type={"text"}
                                                        onChange={(e) => setTlsGenerateConfig(prev => ({ ...prev, common_name: e.currentTarget.value }))}
                                                        value={tlsGenerateConfig.common_name}
                                                        placeholder={"Common Name (CN)"}
                                                        className={"border border-white px-3 py-3 rounded-2xl"}
                                                    />
                                                </div>
                                                <div className={"flex flex-col w-1/2 gap-2"}>
                                                    <p className={"text-sm font-semibold"}>Validity</p>
                                                    <Input
                                                        required
                                                        type={"number"}
                                                        onChange={(e) => setTlsGenerateConfig(prev => ({ ...prev, valid_days: parseInt(e.currentTarget.value) }))}
                                                        value={tlsGenerateConfig.valid_days}
                                                        placeholder={"Validity (days)"}
                                                        className={"border border-white px-3 py-3 rounded-2xl"}
                                                    />
                                                </div>
                                            </div>
                                            <div className={"flex flex-row gap-5"}>
                                                <div className={"flex flex-col w-1/2 gap-2"}>
                                                    <p className={"text-sm font-semibold"}>RSA bits</p>
                                                    <div className={"relative"}>
                                                        <Select
                                                            required
                                                            onChange={(e) => setTlsGenerateConfig(prev => ({ ...prev, rsa_bits: parseInt(e.currentTarget.value) as 2048 | 4096 }))}
                                                            value={tlsGenerateConfig.rsa_bits}
                                                            className={"block w-full rounded-2xl px-3 py-3 appearance-none border border-white"}
                                                        >
                                                            <option
                                                                value={2048}
                                                                className={"bg-stone-900 hover:bg-stone-900/80 text-white"}
                                                            >2048</option>
                                                            <option
                                                                value={4096}
                                                                className={"bg-stone-900 hover:bg-stone-900/80 text-white"}
                                                            >4096</option>
                                                        </Select>
                                                        <ChevronDown size={15} className={"group pointer-events-none absolute right-5 top-5"}/>
                                                    </div>
                                                </div>
                                                <div className={"flex flex-row w-1/2 items-center justify-end"}>
                                                    <div className={"items-center justify-center flex h-full mt-6"}>
                                                        <motion.button
                                                            whileHover={{ scale: 1.05 }}
                                                            whileTap={{ scale: 0.95 }}
                                                            disabled={mutation.loading}
                                                            onClick={generateTLSCertificates}
                                                            className={"px-3 py-2 rounded-full bg-white text-black hover:bg-white/80 transition-colors duration-200 disabled:opacity-70 disabled:cursor-not-allowed"}
                                                        >
                                                            Generate
                                                        </motion.button>
                                                    </div>
                                                </div>
                                            </div>
                                        </div>
                                    </div>
                                </motion.div>
                            )}
                            {!showSetupTLS && (
                                <motion.div
                                    key={"view-tls-configuration"}
                                    initial={{ opacity: 0, y: -10 }}
                                    animate={{ opacity: 1, y: 0 }}
                                    exit={{ opacity: 0, y: -10 }}
                                    transition={{ duration: 0.2 }}
                                    className={"border-box w-full bg-stone-800 rounded-4xl p-5"}
                                >
                                    {proxyServerStatus.loading ? (
                                        <div className={"flex items-center justify-center py-20"}>
                                            <div className={"w-10 h-10 border-4 border-t-white border-stone-800 rounded-full animate-spin"}/>
                                        </div>
                                    ) : (
                                        <div>
                                            {proxyServerStatus.data.config.tls_enabled ? (
                                                <div className={"flex flex-col gap-5"}>
                                                    <div className={"flex flex-row items-center gap-5 justify-between"}>
                                                        <p className={"text-xl font-semibold"}>TLS Configuration</p>
                                                        <div className={"flex items-center justify-center gap-2"}>
                                                            <motion.button
                                                                whileHover={{ scale: 1.05 }}
                                                                whileTap={{ scale: 0.95 }}
                                                                disabled={mutation.loading}
                                                                onClick={() => setShowDeleteTLSConfirmation(true)}
                                                                className={"p-2 rounded-full bg-red-500 hover:bg-red-500/80 text-white disabled:opacity-70 disabled:cursor-not-allowed"}
                                                            >
                                                                <Trash size={15}/>
                                                            </motion.button>
                                                        </div>
                                                    </div>
                                                    <div className={"flex flex-row w-full gap-5"}>
                                                        <div className={"flex items-center justify-center"}>
                                                            <div className={"bg-amber-500 p-2 rounded-full"}>
                                                                <Clock size={15}/>
                                                            </div>
                                                        </div>
                                                        <div className={"flex flex-col"}>
                                                            <p className={"text-md"}>TLS certificate path</p>
                                                            <p className={"text-md font-semibold"}>{proxyServerStatus.data.config.tls_cert_path}</p>
                                                        </div>
                                                    </div>
                                                    <div className={"flex flex-row w-full gap-5"}>
                                                        <div className={"flex items-center justify-center"}>
                                                            <div className={"bg-amber-500 p-2 rounded-full"}>
                                                                <Clock size={15}/>
                                                            </div>
                                                        </div>
                                                        <div className={"flex flex-col"}>
                                                            <p className={"text-md"}>TLS key path</p>
                                                            <p className={"text-md font-semibold"}>{proxyServerStatus.data.config.tls_key_path}</p>
                                                        </div>
                                                    </div>
                                                    <div className={"flex flex-row w-full gap-5"}>
                                                        <div className={"flex items-center justify-center"}>
                                                            <div className={"bg-amber-500 p-2 rounded-full"}>
                                                                <Clock size={15}/>
                                                            </div>
                                                        </div>
                                                        <div className={"flex flex-col"}>
                                                            <p className={"text-md"}>TLS certificate checksum</p>
                                                            <p className={"text-md font-semibold truncate"}>{`SHA-256:${proxyServerStatus.data.config.tls_cert_hash}`}</p>
                                                        </div>
                                                    </div>
                                                    <div className={"flex flex-row w-full gap-5"}>
                                                        <div className={"flex items-center justify-center"}>
                                                            <div className={"bg-amber-500 p-2 rounded-full"}>
                                                                <Clock size={15}/>
                                                            </div>
                                                        </div>
                                                        <div className={"flex flex-col"}>
                                                            <p className={"text-md"}>TLS key checksum</p>
                                                            <p className={"text-md font-semibold truncate"}>{`SHA-256:${proxyServerStatus.data.config.tls_key_hash}`}</p>
                                                        </div>
                                                    </div>
                                                </div>
                                            ) : (
                                                <div className={"flex flex-col gap-5 items-center justify-center py-5"}>
                                                    <p className={"text-md font-semibold"}>TLS is not set up yet, server listening on HTTP</p>
                                                    <div className={"flex flex-row justify-center items-center"}>
                                                        <motion.button
                                                            whileHover={{ scale: 1.05 }}
                                                            whileTap={{ scale: 0.95 }}
                                                            disabled={mutation.loading}
                                                            onClick={() => setShowSetupTLS(true)}
                                                            className={"px-3 py-1 rounded-full bg-white text-black hover:bg-white/80 transition-colors duration-200"}
                                                        >
                                                            Setup TLS
                                                        </motion.button>
                                                    </div>
                                                </div>
                                            )}
                                        </div>
                                    )}
                                </motion.div>
                            )}
                        </AnimatePresence>
                    </div>
                </div>
            </div>
        </PageLayout>
    );
}
