import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";

import {
    generateIncidentAnalysis,
    getIncidentById,
    type IncidentAnalysis,
} from "../api/incidents";
import type { Incident } from "../types/incident";

export function IncidentDetailPage() {
    const { id } =
        useParams();

    const [incident, setIncident] =
        useState<Incident | null>(
            null,
        );

    const [loading, setLoading] =
        useState(true);

    const [error, setError] =
        useState<string | null>(null);
    const [analysis, setAnalysis] =
        useState<IncidentAnalysis | null>(
            null,
        );

    const [analysisLoading, setAnalysisLoading] =
        useState(false);

    const [analysisError, setAnalysisError] =
        useState<string | null>(null);
    function formatDate(
        value: string | null,
    ): string {
        if (value === null) {
            return "Not resolved";
        }

        return new Date(
            value,
        ).toLocaleString();
    }
    async function handleGenerateAnalysis() {
        if (incident === null) {
            return;
        }

        setAnalysisLoading(true);
        setAnalysisError(null);

        try {
            const data =
                await generateIncidentAnalysis(
                    incident.id,
                );

            setAnalysis(data);
        } catch {
            setAnalysisError(
                "Failed to generate incident analysis.",
            );
        } finally {
            setAnalysisLoading(false);
        }
    }
    useEffect(() => {
        async function loadIncident() {
            const parsedID =
                Number(id);

            if (
                !Number.isInteger(
                    parsedID,
                ) ||
                parsedID <= 0
            ) {
                setError(
                    "Invalid incident ID.",
                );

                setLoading(false);
                return;
            }

            try {
                const data =
                    await getIncidentById(
                        parsedID,
                    );

                setIncident(data);
            } catch {
                setError(
                    "Failed to load incident.",
                );
            } finally {
                setLoading(false);
            }
        }


        void loadIncident();
    }, [id]);


    if (loading) {
        return (
            <main>
                <p>
                    Loading incident...
                </p>
            </main>
        );
    }

    if (
        error ||
        incident === null
    ) {
        return (
            <main>
                <h1>Incident</h1>

                <p>
                    {error ??
                        "Incident not found."}
                </p>

                <Link to="/incidents">
                    Back to incidents
                </Link>
            </main>
        );
    }

    return (
        <main>
            <Link to="/incidents">
                ← Back to incidents
            </Link>

            <h1>
                Incident #{incident.id}
            </h1>

            <article>
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
                    Failure count:{" "}
                    {incident.failureCount}
                </p>

                <p>
                    Started at:{" "}
                    {formatDate(
                        incident.startedAt,
                    )}
                </p>

                <p>
                    Resolved at:{" "}
                    {formatDate(
                        incident.resolvedAt,
                    )}
                </p>

                <p>
                    Last error:{" "}
                    {incident.lastErrorMessage ??
                        "None"}
                </p>
            </article>
            <section>
                <h2>
                    AI Incident Analysis
                </h2>

                <p>
                    Generate an incident summary,
                    possible causes, and recommended
                    remediation actions.
                </p>

                <button
                    type="button"
                    onClick={() =>
                        void handleGenerateAnalysis()
                    }
                    disabled={analysisLoading}
                >
                    {analysisLoading
                        ? "Generating..."
                        : "Generate Analysis"}
                </button>

                {analysisError && (
                    <p>
                        {analysisError}
                    </p>
                )}

                {analysis && (
                    <div>
                        <h3>Summary</h3>

                        <p>
                            {analysis.summary}
                        </p>

                        <h3>
                            Possible Causes
                        </h3>

                        {analysis.possibleCauses.length ===
                            0 ? (
                            <p>
                                No possible causes returned.
                            </p>
                        ) : (
                            <ul>
                                {analysis.possibleCauses.map(
                                    (cause) => (
                                        <li key={cause}>
                                            {cause}
                                        </li>
                                    ),
                                )}
                            </ul>
                        )}

                        <h3>
                            Recommended Actions
                        </h3>

                        {analysis.recommendedActions.length ===
                            0 ? (
                            <p>
                                No recommended actions returned.
                            </p>
                        ) : (
                            <ul>
                                {analysis.recommendedActions.map(
                                    (action) => (
                                        <li key={action}>
                                            {action}
                                        </li>
                                    ),
                                )}
                            </ul>
                        )}
                    </div>
                )}
            </section>
        </main>
    );
}