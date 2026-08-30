import { useEffect, useState } from "react";

import { apiClient } from "../api/client";

type BackendStatus = "checking" | "online" | "offline";

export function HomePage() {
    const [backendStatus, setBackendStatus] =
        useState<BackendStatus>("checking");

    useEffect(() => {
        async function checkBackend() {
            try {
                await apiClient.get("/health");
                setBackendStatus("online");
            } catch {
                setBackendStatus("offline");
            }
        }

        void checkBackend();
    }, []);

    return (
        <main>
            <h1>CloudPulse</h1>

            <p>
                Cloud Service Monitoring &
                AI-Assisted Incident Platform
            </p>

            <section>
                <h2>System Status</h2>

                <p>
                    Backend:{" "}
                    <strong>
                        {backendStatus}
                    </strong>
                </p>
            </section>

            <section>
                <h2>Platform Capabilities</h2>

                <ul>
                    <li>
                        Concurrent service health monitoring
                    </li>

                    <li>
                        Automatic incident detection and recovery
                    </li>

                    <li>
                        Prometheus observability metrics
                    </li>

                    <li>
                        AI-assisted incident analysis
                    </li>
                </ul>
            </section>
        </main>
    );
}