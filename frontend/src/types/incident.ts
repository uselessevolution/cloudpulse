export type IncidentStatus =
    | "OPEN"
    | "RESOLVED";

export interface Incident {
    id: number;
    serviceId: number;
    status: IncidentStatus;
    startedAt: string;
    resolvedAt: string | null;
    failureCount: number;
    lastErrorMessage: string | null;
    createdAt: string;
    updatedAt: string;
}