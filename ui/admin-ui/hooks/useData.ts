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

import {useRouter} from "next/router";
import {useEffect, useState, useCallback} from "react";

interface Data<T>{
    data: T;
    loading: boolean;
    error: Error | null;
    refetch: () => Promise<void>;
}

export function useData<T>(path: string, key: string) : Data<T> {
    const [data, setData] = useState<T>(null as any);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<Error | null>(null);

    const fetchData = useCallback(
        async() => {
            setLoading(true);
            setError(null);

            try{
                const response = await fetch(path, { method: "GET" });
                if (!response.ok) {
                    let errMsg = response.statusText || "Failed to fetch data";
                    try{
                        const errorData = await response.json();
                        errMsg = errorData.error.details || errMsg;
                    } catch(err) {
                        // Ignore JSON parsing errors and use the default message
                    }
                    throw new Error(errMsg);
                }
                const responseData = await response.json();
                setData(responseData[key]);
            } catch (err: unknown) {
                setError(err instanceof Error ? err : new Error("An unknown error occurred"));
            } finally {
                setLoading(false);
            }
        }, [path, key]
    );

    useEffect(() => {
        fetchData();
    }, [fetchData]);

    return { data, loading, error, refetch: fetchData };
}