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

import React from "react";
import {AnimatePresence, motion} from "motion/react";
import {X} from "lucide-react";

interface ModalDialogProps {
    title: string;
    visible: boolean;
    onClose: () => void;
    children: React.ReactNode;
}

export function ModalDialog({title, visible, onClose, children}: ModalDialogProps) {
    return(
        <AnimatePresence mode={"popLayout"}>
            {visible && (
                <motion.div
                    key={"visible"}
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1 }}
                    exit={{ opacity: 0 }}
                    className={"backdrop-blur-2xl fixed inset-0 flex items-start justify-center z-50"}
                >
                    <motion.div
                        initial={{ opacity: 0, y: -20 }}
                        animate={{ opacity: 1, y: 0 }}
                        exit={{ opacity: 0, y: -20 }}
                        transition={{ duration: 0.3 }}
                        className={"fixed top-20 left-1/2 transform -translate-x-1/2 w-full md:w-1/2 lg:w-1/3 bg-stone-800 rounded-4xl z-50 border-box p-5"}
                    >
                        <div className={"flex flex-col gap-5"}>
                            <div className={"flex flex-row items-center gap-5 justify-between"}>
                                <p className={"text-xl font-semibold"}>{title}</p>
                                <div className={"flex items-center justify-center"}>
                                    <motion.button
                                        whileHover={{ scale: 1.05 }}
                                        whileTap={{ scale: 0.95 }}
                                        onClick={onClose}
                                        className={"p-2 rounded-full bg-white text-black"}
                                    >
                                        <X size={15}/>
                                    </motion.button>
                                </div>
                            </div>
                        </div>
                        {children}
                    </motion.div>
                </motion.div>
            )}
        </AnimatePresence>
    );
}