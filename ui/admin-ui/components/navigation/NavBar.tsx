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

type NavLink = {
    label: string;
    href: string;
    onClick?: (e: React.MouseEvent) => void;
};

interface NavBarProps {
    title?: string;
    links: NavLink[];
    rightContent?: React.ReactNode;
    className?: string;
}

export default function NavBar({ title, links, rightContent, className }: NavBarProps) {
    return (
        <nav className={`w-full bg-stone-950 border-b border-stone-800 ${className}`}>
            <div className={`max-w-350 mx-auto px-4 sm:px-6 lg:px-8`}>
                <div className={`flex h-16 items-center justify-between`}>
                    <div className={`flex items-center gap-4`}>
                        <div className={`text-lg font-mono font-semibold text-white`}>{title}</div>

                        <div className={`hidden md:flex items-center gap-4`}>
                            {links.map((link) => (
                                <a
                                    key={link.href}
                                    href={link.href}
                                    onClick={link.onClick}
                                    className={`px-3 py-1 rounded text-sm text-white hover:bg-stone-800`}
                                >
                                    {link.label}
                                </a>
                            ))}
                        </div>
                    </div>

                    <div className={`flex items-center gap-2`}>{rightContent}</div>
                </div>
            </div>
        </nav>
    );
}