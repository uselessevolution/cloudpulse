export type RuntimeStatus =
    | "HEALTHY"
    | "DEGRADED"
    | "DOWN";

export interface Service {
    id: number;
    name: string;
    url: string;
    expectedStatus: number;
    checkIntervalSeconds: number;
    timeoutSeconds: number;
    enabled: boolean;
    runtimeStatus: RuntimeStatus;
    consecutiveFailures: number;
    consecutiveSuccesses: number;
    lastCheckedAt: string | null;
    createdAt: string;
    updatedAt: string;
}