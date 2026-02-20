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
import Image from "next/image";
import {motion,AnimatePresence} from "motion/react";

type NavLink = {
    label: string;
    href: string;
    active: boolean;
};

interface NavBarProps {
    links: NavLink[];
}

export default function NavBar({ links }: NavBarProps) {
    return (
        <nav className={"w-full py-2"}>
            <div className={"flex flex-row justify-between gap-5 w-full items-center h-15"}>
                <div className={"px-5 flex items-center justify-center h-full bg-stone-800 rounded-4xl"}>
                    <a href={"/"} className={"text-2xl font-semibold flex flex-row"}>Wasm<p className={"text-2xl font-semibold text-amber-500"}>Forge</p></a>
                </div>
                <div className={"px-2 flex flex-row items-center gap-5 h-full bg-stone-800 rounded-4xl min-w-2/3"}>
                    <AnimatePresence mode={"sync"}>
                        {links.map((link, idx) => (
                            <motion.a
                                key={link.label}
                                className={`px-4 py-2 text-lg rounded-4xl ${link.active ? "bg-stone-600 text-amber-500" : "hover:bg-stone-600 hover:text-amber-500 transition-colors duration-200"}`}
                                href={link.href}
                                initial={{ opacity: 0, y: -10 }}
                                animate={{ opacity: 1, y: 0 }}
                                exit={{ opacity: 0, y: -10 }}
                                transition={{ duration: 0.2 * idx }}
                            >
                                {link.label}
                            </motion.a>
                        ))}
                    </AnimatePresence>
                </div>
                <div className={"flex items-center justify-center h-full bg-stone-800 rounded-full px-5"}>
                    <a href={"https://github.com/mikhail5545/wasmforge"}><Image width={20} height={20} src={"/GitHub_Invertocat_White.svg"} alt={"GitHub"}/></a>
                </div>
            </div>
        </nav>
    );
}