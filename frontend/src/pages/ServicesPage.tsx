import { useEffect, useState } from "react";

import { getServices } from "../api/services";
import type { Service } from "../types/service";

export function ServicesPage() {
    const [services, setServices] =
        useState<Service[]>([]);

    const [loading, setLoading] =
        useState(true);

    const [error, setError] =
        useState<string | null>(null);

    useEffect(() => {
        async function loadServices() {
            try {
                const data =
                    await getServices();

                setServices(data);
            } catch {
                setError(
                    "Failed to load services.",
                );
            } finally {
                setLoading(false);
            }
        }

        void loadServices();
    }, []);

    if (loading) {
        return (
            <main>
                <h1>Services</h1>
                <p>Loading services...</p>
            </main>
        );
    }

    if (error) {
        return (
            <main>
                <h1>Services</h1>
                <p>{error}</p>
            </main>
        );
    }
    const healthyCount =
        services.filter(
            (service) =>
                service.runtimeStatus ===
                "HEALTHY",
        ).length;

    const degradedCount =
        services.filter(
            (service) =>
                service.runtimeStatus ===
                "DEGRADED",
        ).length;

    const downCount =
        services.filter(
            (service) =>
                service.runtimeStatus ===
                "DOWN",
        ).length;
    return (
        <main>
            <h1>Services</h1>
            <section>
                <p>
                    Total:{" "}
                    <strong>
                        {services.length}
                    </strong>
                </p>

                <p>
                    Healthy:{" "}
                    <strong>
                        {healthyCount}
                    </strong>
                </p>

                <p>
                    Degraded:{" "}
                    <strong>
                        {degradedCount}
                    </strong>
                </p>

                <p>
                    Down:{" "}
                    <strong>
                        {downCount}
                    </strong>
                </p>
            </section>
            {services.length === 0 ? (
                <p>No monitored services.</p>
            ) : (
                <div>
                    {services.map(
                        (service) => (
                            <article
                                key={service.id}
                            >
                                <h2>
                                    {service.name}
                                </h2>

                                <p>
                                    Status:{" "}
                                    <strong>
                                        {
                                            service.runtimeStatus
                                        }
                                    </strong>
                                </p>

                                <p>
                                    URL:{" "}
                                    {service.url}
                                </p>

                                <p>
                                    Expected status:{" "}
                                    {
                                        service.expectedStatus
                                    }
                                </p>

                                <p>
                                    Check interval:{" "}
                                    {
                                        service.checkIntervalSeconds
                                    }
                                    s
                                </p>

                                <p>
                                    Timeout:{" "}
                                    {
                                        service.timeoutSeconds
                                    }
                                    s
                                </p>

                                <p>
                                    Enabled:{" "}
                                    {service.enabled
                                        ? "Yes"
                                        : "No"}
                                </p>

                                <p>
                                    Last checked:{" "}
                                    {
                                        service.lastCheckedAt ??
                                        "Never"
                                    }
                                </p>
                            </article>
                        ),
                    )}
                </div>
            )}
        </main>
    );
}