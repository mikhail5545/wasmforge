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

import {motion} from "motion/react";
import React from "react";

interface PluginGridListCardProps {
    plugin: WasmForge.Plugin;
    index: number;
    onClick: (() => void) | null;
    currentlySelected?: boolean;
}

interface RouteGridListCardProps {
    route: WasmForge.Route;
    index: number;
    onClick: (() => void) | null;
    currentlySelected?: boolean;
}

interface RoutePluginGridListCardProps {
    routePlugin: WasmForge.RoutePlugin;
    index: number;
    onClick: (() => void) | null;
    currentlySelected?: boolean;
}

interface GridListCardProps {
    content: React.ReactNode;
    index: number;
    onClick: (() => void) | null;
    currentlySelected?: boolean;
}

export function GridListCard({ content, index, onClick, currentlySelected }: GridListCardProps) {
    if (onClick) {
        return (
            <motion.div
                initial={{ opacity: 0, y: -20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.2, delay: index * 0.05 }}
                onClick={onClick}
                className={`col-span-1 w-full rounded bg-stone-800 p-3 cursor-pointer 
                ${currentlySelected ? "ring-2 ring-white ring-offset-2" : "hover:ring-2 hover:ring-white hover:ring-offset-2"}`}
            >
                {content}
            </motion.div>
        );
    }
    return (
        <motion.div
            initial={{ opacity: 0, y: -20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.2, delay: index * 0.05 }}
            className={"col-span-1 rounded bg-stone-800 p-3"}
        >
            {content}
        </motion.div>
    );
}

export function PluginGridListCard({ plugin, index, onClick, currentlySelected }: PluginGridListCardProps) {
    const cardContent = (
        <div className={"flex flex-row"}>
            <div className={"flex flex-col"}>
                <p className={"text-sm text-stone-400"}>Name</p>
                <p className={"text-md font-semibold"}>{plugin.name}</p>
            </div>
            <div className={"flex flex-col"}>
                <p className={"text-sm text-stone-400"}>Filename</p>
                <p className={"text-md font-semibold"}>{plugin.filename}</p>
            </div>
        </div>
    );
    return <GridListCard content={cardContent} index={index} onClick={onClick} currentlySelected={currentlySelected} />;
}

export function RouteGridListCard({ route, index, onClick, currentlySelected }: RouteGridListCardProps) {
    const cardContent = (
        <div className={"flex flex-row"}>
            <div className={"flex flex-col"}>
                <p className={"text-sm text-stone-400"}>Path</p>
                <p className={"text-md font-semibold"}>{route.path}</p>
            </div>
            <div className={"flex flex-col"}>
                <p className={"text-sm text-stone-400"}>Target URL</p>
                <p className={"text-md font-semibold"}>{route.target_url}</p>
            </div>
        </div>
    );
    return <GridListCard content={cardContent} index={index} onClick={onClick} currentlySelected={currentlySelected} />;
}

export function RoutePluginGridListCard({ routePlugin, index, onClick, currentlySelected }: RoutePluginGridListCardProps) {
    const cardContent = (
        <div className={"flex flex-row px-3"}>
            <div className={"flex flex-col w-1/2"}>
                <p className={"text-sm text-stone-400"}>Plugin name</p>
                <p className={"text-md font-semibold"}>{routePlugin.plugin?.name || "N/A"}</p>
            </div>
            <div className={"flex flex-col w-1/2"}>
                <p className={"text-sm text-stone-400"}>Execution order</p>
                <p className={"text-md font-semibold"}>{routePlugin.execution_order}</p>
            </div>
        </div>
    );
    return <GridListCard content={cardContent} index={index} onClick={onClick} currentlySelected={currentlySelected} />;
}