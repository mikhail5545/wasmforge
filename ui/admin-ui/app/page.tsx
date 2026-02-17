'use client';

import NavBar from "@/components/navigation/NavBar";
import React from "react";
import {motion} from "motion/react";
import {
    MoveRight,
    ChevronRight,
} from "lucide-react";

import {usePaginatedData} from "@/hooks/usePaginatedData";
import {ErrorDialog} from "@/components/dialog/ErrorDialog";

export default function Home() {
    const links = [
        { label: "Dashboard", href: "/dashboard" },
        { label: "Routes", href: "/routes" },
        { label: "Plugins", href: "/plugins" },
    ];

    const routePaginatedData = usePaginatedData<WasmForge.Route>(
        "/api/routes?e=true",
        "routes",
        10,
        "created_at",
        "desc",
        { preload: true }
        );

    return (
        <div className={"flex min-h-screen bg-stone-950 font-mono text-white"}>
            <div className={"flex flex-col w-full"}>
            </div>
        </div>
    );
}