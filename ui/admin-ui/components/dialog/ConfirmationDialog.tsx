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

import {motion,AnimatePresence} from "motion/react";
import {X} from "lucide-react";
import React from "react";

interface ConfirmationDialogProps {
    title: string;
    message: string;
    isOpen: boolean;
    accentColor?: string;
    onConfirm: () => void;
    onCancel: () => void;
}

export function ConfirmationDialog({ title, message, isOpen, accentColor, onConfirm, onCancel }: ConfirmationDialogProps) {
    return(
        <AnimatePresence mode={"wait"}>
            {isOpen && (
                <motion.div
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1 }}
                    exit={{ opacity: 0 }}
                    className={"backdrop-blur-2xl fixed inset-0 flex items-start justify-center z-50"}
                >
                    <motion.div
                        key={"route-selection"}
                        initial={{ opacity: 0, y: -20 }}
                        animate={{ opacity: 1, y: 0 }}
                        exit={{ opacity: 0, y: -20 }}
                        transition={{ duration: 0.3 }}
                        className={"fixed top-20 left-1/2 transform -translate-x-1/2 w-full md:w-1/2 lg:w-1/3 bg-stone-900/80 rounded-md z-50 p-5"}
                    >
                        <div className={"flex flex-col gap-5"}>
                            <div className={"flex flex-row items-center justify-start gap-2"}>
                                <motion.button
                                    whileHover={{ scale: 1.05 }}
                                    whileTap={{ scale: 0.95 }}
                                    onClick={onCancel}
                                    className={"bg-stone-800 text-sm font-semibold px-2 py-2 rounded disabled:opacity-50 flex items-center justify-center gap-2 hover:bg-stone-700/80 transition-colors duration-200"}
                                >
                                    <X size={12}/>
                                </motion.button>
                                <p className={`text-lg font-semibold ${accentColor ? `text-${accentColor}` : "text-white"}`}>{title}</p>
                            </div>
                            <div className={"flex flex-col"}>
                                <p className={"text-md"}>{message}</p>
                            </div>
                            <div className={"flex flex-row gap-5 mt-5 justify-between"}>
                                <motion.button
                                    whileHover={{ scale: 1.05 }}
                                    whileTap={{ scale: 0.95 }}
                                    onClick={onCancel}
                                    className={"border border-stone-800 text-sm font-semibold px-4 py-2 rounded disabled:opacity-50 flex items-center justify-center gap-2 hover:bg-stone-700/80 transition-colors duration-200"}
                                >
                                    Cancel
                                </motion.button>
                                <motion.button
                                    whileHover={{ scale: 1.05 }}
                                    whileTap={{ scale: 0.95 }}
                                    onClick={onConfirm}
                                    className={`bg-red-500 hover:bg-red-500/80 text-sm font-semibold px-4 py-2 rounded disabled:opacity-50 flex items-center justify-center gap-2 transition-colors duration-200`}
                                >
                                    Proceed
                                </motion.button>
                            </div>
                        </div>
                    </motion.div>
                </motion.div>
            )}
        </AnimatePresence>
    );
}