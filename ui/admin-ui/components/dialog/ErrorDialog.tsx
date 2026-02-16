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

import {AnimatePresence,motion} from "motion/react";

interface ErrorDialogProps {
    title: string;
    message: string;
    isOpen: boolean;
    onClose: () => void;
}

export function ErrorDialog({ title, message, isOpen, onClose }: ErrorDialogProps) {
    return(
        <AnimatePresence>
            {isOpen && (
                <motion.div
                    key={"modal"}
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1 }}
                    exit={{ opacity: 0 }}
                    className={"fixed inset-0 flex items-center justify-center z-50 backdrop-blur-xl"}
                >
                    <div className={"bg-stone-800 rounded-lg p-6 w-full max-w-md mx-4"}>
                        <h2 className={"text-xl font-semibold mb-4 text-red-500"}>{title}</h2>
                        <p className={"text-stone-400 mb-6"}>{message}</p>
                        <button
                            onClick={onClose}
                            className={"px-4 py-2 bg-red-700/70 rounded text-sm w-full"}
                        >
                            Close
                        </button>
                    </div>
                </motion.div>
            )}
        </AnimatePresence>
    );
}