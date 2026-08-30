import { useEffect, useState } from "react";
import { Link } from "react-router-dom";

import { getIncidents } from "../api/incidents";
import type { Incident } from "../types/incident";

export function IncidentsPage() {
    const [incidents, setIncidents] =
        useState<Incident[]>([]);

    const [loading, setLoading] =
        useState(true);

    const [error, setError] =
        useState<string | null>(null);

    useEffect(() => {
        async function loadIncidents() {
            try {
                const data =
                    await getIncidents();

                setIncidents(data);
            } catch {
                setError(
                    "Failed to load incidents.",
                );
            } finally {
                setLoading(false);
            }
        }

        void loadIncidents();
    }, []);

    if (loading) {
        return (
            <main>
                <h1>Incidents</h1>
                <p>Loading incidents...</p>
            </main>
        );
    }

    if (error) {
        return (
            <main>
                <h1>Incidents</h1>
                <p>{error}</p>
            </main>
        );
    }

    const openCount =
        incidents.filter(
            (incident) =>
                incident.status === "OPEN",
        ).length;

    const resolvedCount =
        incidents.filter(
            (incident) =>
                incident.status === "RESOLVED",
        ).length;

    return (
        <main>
            <h1>Incidents</h1>

            <section>
                <p>
                    Total:{" "}
                    <strong>
                        {incidents.length}
                    </strong>
                </p>

                <p>
                    Open:{" "}
                    <strong>
                        {openCount}
                    </strong>
                </p>

                <p>
                    Resolved:{" "}
                    <strong>
                        {resolvedCount}
                    </strong>
                </p>
            </section>

            {incidents.length === 0 ? (
                <p>No incidents recorded.</p>
            ) : (
                <div>
                    {incidents.map(
                        (incident) => (
                            <article
                                key={incident.id}
                            >
                                <h2>
                                    Incident #{incident.id}
                                </h2>

                                <p>
                                    Status:{" "}
                                    <strong>
                                        {incident.status}
                                    </strong>
                                </p>

                                <p>
                                    Service ID:{" "}
                                    {incident.serviceId}
                                </p>

                                <p>
                                    Failures:{" "}
                                    {incident.failureCount}
                                </p>

                                <p>
                                    Started:{" "}
                                    {incident.startedAt}
                                </p>

                                <p>
                                    Resolved:{" "}
                                    {incident.resolvedAt ??
                                        "Not resolved"}
                                </p>

                                <p>
                                    Last error:{" "}
                                    {incident.lastErrorMessage ??
                                        "None"}
                                </p>

                                <Link
                                    to={`/incidents/${incident.id}`}
                                >
                                    View Incident
                                </Link>
                            </article>
                        ),
                    )}
                </div>
            )}
        </main>
    );
}