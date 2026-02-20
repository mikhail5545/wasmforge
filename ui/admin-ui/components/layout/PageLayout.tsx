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

import { Quicksand } from "next/font/google";
import React, {Suspense} from "react";

const quicksand = Quicksand({
    subsets: ["latin"],
    weight: ["400", "500", "600", "700"],
});

export default function PageLayout({ children }: { children: React.ReactNode }) {
    return (
        <Suspense fallback={
            <div className={`flex min-h-screen bg-stone-950 font-mono text-white ${quicksand.className}`}>
                <div className={`flex flex-col w-full`}>
                    <div className={`flex flex-col px-5 md:px-10 lg:px-20 xl:px-40 py-5 gap-5`}>
                        <div className={"flex flex-col w-full py-40 items-center justify-center"}>
                            <div className={"w-10 h-10 border-4 border-t-white border-stone-600 rounded-full animate-spin"}/>
                        </div>
                    </div>
                </div>
            </div>
        }>
            <div className={`flex min-h-screen bg-stone-950 font-mono text-white ${quicksand.className}`}>
                <div className={`flex flex-col w-full`}>
                    <div className={`flex flex-col px-5 md:px-10 lg:px-20 xl:px-40 py-5 gap-5`}>
                        {children}
                    </div>
                </div>
            </div>
        </Suspense>
    );
}