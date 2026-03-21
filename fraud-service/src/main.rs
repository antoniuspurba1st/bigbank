use actix_web::middleware::Logger;
use actix_web::{App, HttpRequest, HttpResponse, HttpServer, get, post, web};
use chrono::Utc;
use log::info;
use prometheus::{Counter, CounterVec, Histogram, HistogramVec, Registry, TextEncoder, Encoder};
use serde::{Deserialize, Serialize};
use std::env;
use std::sync::Mutex;
use uuid::Uuid;

const CORRELATION_ID_HEADER: &str = "X-Correlation-Id";
const DEFAULT_FRAUD_REJECTION_LIMIT: f64 = 1_000_000.0;

// Phase 8: Prometheus metrics registry
struct MetricsRegistry {
    request_count: CounterVec,
    error_count: CounterVec,
    request_duration: HistogramVec,
}

impl MetricsRegistry {
    fn new(registry: &Registry) -> Result<Self, prometheus::Error> {
        let request_count = CounterVec::new(
            prometheus::Opts::new("http_requests_total", "Total number of HTTP requests"),
            &["method", "endpoint", "status"],
        )?;
        registry.register(Box::new(request_count.clone()))?;

        let error_count = CounterVec::new(
            prometheus::Opts::new("http_errors_total", "Total number of HTTP errors"),
            &["method", "endpoint", "status"],
        )?;
        registry.register(Box::new(error_count.clone()))?;

        let request_duration = HistogramVec::new(
            prometheus::HistogramOpts::new("http_request_duration_seconds", "HTTP request duration in seconds"),
            &["method", "endpoint"],
        )?;
        registry.register(Box::new(request_duration.clone()))?;

        Ok(MetricsRegistry {
            request_count,
            error_count,
            request_duration,
        })
    }
}


#[derive(Deserialize)]
struct FraudRequest {
    reference: Option<String>,
    from_account: Option<String>,
    to_account: Option<String>,
    amount: Option<f64>,
}

#[derive(Serialize)]
struct ApiResponse<T> {
    status: String,
    message: String,
    correlation_id: String,
    data: Option<T>,
}

#[derive(Serialize)]
struct ApiError {
    status: String,
    code: String,
    message: String,
    correlation_id: String,
}

#[derive(Serialize)]
struct FraudDecision {
    decision: String,
    approved: bool,
    reason: String,
    checked_at: String,
}

struct NormalizedFraudRequest {
    reference: String,
    from_account: String,
    to_account: String,
    amount: f64,
}

#[get("/health")]
async fn health(request: HttpRequest) -> HttpResponse {
    let correlation_id = correlation_id(&request);

    HttpResponse::Ok()
        .insert_header((CORRELATION_ID_HEADER, correlation_id.clone()))
        .json(ApiResponse {
            status: "success".to_string(),
            message: "Fraud service is healthy".to_string(),
            correlation_id,
            data: Some(serde_json::json!({
                "service": "fraud",
                "status": "UP",
            })),
        })
}

#[post("/fraud/check")]
async fn fraud_check(request: HttpRequest, body: web::Json<FraudRequest>) -> HttpResponse {
    let correlation_id = correlation_id(&request);
    let normalized = match normalize_request(body.into_inner()) {
        Ok(value) => value,
        Err((code, message)) => {
            return HttpResponse::BadRequest()
                .insert_header((CORRELATION_ID_HEADER, correlation_id.clone()))
                .json(ApiError {
                    status: "error".to_string(),
                    code: code.to_string(),
                    message: message.to_string(),
                    correlation_id,
                });
        }
    };

    info!(
        "correlation_id={} event=fraud_check reference={} from_account={} to_account={} amount={}",
        correlation_id,
        normalized.reference,
        normalized.from_account,
        normalized.to_account,
        normalized.amount,
    );

    let threshold = fraud_rejection_limit();
    let decision = if normalized.amount > threshold {
        FraudDecision {
            decision: "rejected".to_string(),
            approved: false,
            reason: format!("Amount exceeds fraud threshold {:.2}", threshold),
            checked_at: Utc::now().to_rfc3339(),
        }
    } else {
        FraudDecision {
            decision: "approved".to_string(),
            approved: true,
            reason: "Transaction passed fraud rules".to_string(),
            checked_at: Utc::now().to_rfc3339(),
        }
    };

    let status = if decision.approved {
        "success"
    } else {
        "rejected"
    };
    let message = if decision.approved {
        "Fraud check approved"
    } else {
        "Fraud check rejected"
    };

    info!(
        "correlation_id={} event=fraud_decided reference={} decision={} approved={} reason=\"{}\"",
        correlation_id,
        normalized.reference,
        decision.decision,
        decision.approved,
        decision.reason,
    );

    HttpResponse::Ok()
        .insert_header((CORRELATION_ID_HEADER, correlation_id.clone()))
        .json(ApiResponse {
            status: status.to_string(),
            message: message.to_string(),
            correlation_id,
            data: Some(decision),
        })
}

fn normalize_request(
    request: FraudRequest,
) -> Result<NormalizedFraudRequest, (&'static str, &'static str)> {
    let reference = request.reference.unwrap_or_default().trim().to_string();
    let from_account = request.from_account.unwrap_or_default().trim().to_string();
    let to_account = request.to_account.unwrap_or_default().trim().to_string();
    let amount = request.amount.unwrap_or(0.0);

    if reference.is_empty() {
        return Err(("INVALID_REFERENCE", "Reference is required"));
    }
    if reference.len() > 128 {
        return Err(("INVALID_REFERENCE", "Reference is too long"));
    }
    if !is_valid_reference(&reference) {
        return Err((
            "INVALID_REFERENCE",
            "Reference contains unsupported characters",
        ));
    }
    if from_account.is_empty() || to_account.is_empty() {
        return Err(("INVALID_ACCOUNT", "Both accounts are required"));
    }
    if !is_valid_account(&from_account) || !is_valid_account(&to_account) {
        return Err(("INVALID_ACCOUNT", "Account format is invalid"));
    }
    if from_account == to_account {
        return Err((
            "SAME_ACCOUNT_TRANSFER",
            "Source and destination accounts must differ",
        ));
    }
    if amount <= 0.0 {
        return Err(("INVALID_AMOUNT", "Amount must be greater than zero"));
    }

    Ok(NormalizedFraudRequest {
        reference,
        from_account,
        to_account,
        amount,
    })
}

fn correlation_id(request: &HttpRequest) -> String {
    request
        .headers()
        .get(CORRELATION_ID_HEADER)
        .and_then(|value| value.to_str().ok())
        .map(str::trim)
        .filter(|value| !value.is_empty())
        .map(ToOwned::to_owned)
        .unwrap_or_else(|| Uuid::new_v4().to_string())
}

fn is_valid_reference(reference: &str) -> bool {
    let mut chars = reference.chars();
    let Some(first) = chars.next() else {
        return false;
    };

    if !first.is_ascii_alphanumeric() {
        return false;
    }

    chars.all(|character| character.is_ascii_alphanumeric() || character == '_' || character == '-')
}

fn is_valid_account(account: &str) -> bool {
    let length = account.len();
    if !(3..=64).contains(&length) {
        return false;
    }

    account
        .chars()
        .all(|character| character.is_ascii_alphanumeric() || character == '-')
}

fn fraud_rejection_limit() -> f64 {
    env::var("FRAUD_REJECTION_LIMIT")
        .ok()
        .and_then(|value| value.parse::<f64>().ok())
        .filter(|value| *value > 0.0)
        .unwrap_or(DEFAULT_FRAUD_REJECTION_LIMIT)
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    env_logger::init_from_env(env_logger::Env::new().default_filter_or("info"));
    let port = env::var("PORT").unwrap_or_else(|_| "8082".to_string());
    let threshold = fraud_rejection_limit();
    info!(
        "event=fraud_service_started port={} fraud_rejection_limit={}",
        port, threshold
    );

    HttpServer::new(|| {
        App::new()
            .wrap(Logger::default())
            .service(health)
            .service(fraud_check)
    })
    .bind(("127.0.0.1", port.parse::<u16>().unwrap_or(8082)))?
    .run()
    .await
}

#[cfg(test)]
mod tests {
    use super::*;
    use actix_web::{http::StatusCode, test};

    #[actix_web::test]
    async fn approves_small_amount() {
        let app = test::init_service(App::new().service(fraud_check)).await;
        let request = test::TestRequest::post()
            .uri("/fraud/check")
            .insert_header((CORRELATION_ID_HEADER, "corr-approved"))
            .set_json(serde_json::json!({
                "reference": "ref-1",
                "from_account": "ACC-001",
                "to_account": "ACC-002",
                "amount": 99.50
            }))
            .to_request();

        let response = test::call_service(&app, request).await;

        assert_eq!(response.status(), StatusCode::OK);
    }

    #[actix_web::test]
    async fn rejects_large_amount() {
        let app = test::init_service(App::new().service(fraud_check)).await;
        let request = test::TestRequest::post()
            .uri("/fraud/check")
            .set_json(serde_json::json!({
                "reference": "ref-2",
                "from_account": "ACC-001",
                "to_account": "ACC-002",
                "amount": 2000000.00
            }))
            .to_request();

        let response = test::call_service(&app, request).await;
        let body: serde_json::Value = test::read_body_json(response).await;

        assert_eq!(body["status"], "rejected");
        assert_eq!(body["data"]["decision"], "rejected");
    }

    #[actix_web::test]
    async fn rejects_invalid_payload() {
        let app = test::init_service(App::new().service(fraud_check)).await;
        let request = test::TestRequest::post()
            .uri("/fraud/check")
            .set_json(serde_json::json!({
                "reference": "ref-3",
                "from_account": "ACC-001",
                "to_account": "ACC-001",
                "amount": 10.00
            }))
            .to_request();

        let response = test::call_service(&app, request).await;
        let status = response.status();
        let body: serde_json::Value = test::read_body_json(response).await;

        assert_eq!(status, StatusCode::BAD_REQUEST);
        assert_eq!(body["code"], "SAME_ACCOUNT_TRANSFER");
    }

    #[actix_web::test]
    async fn rejects_invalid_reference_characters() {
        let app = test::init_service(App::new().service(fraud_check)).await;
        let request = test::TestRequest::post()
            .uri("/fraud/check")
            .set_json(serde_json::json!({
                "reference": "bad ref!",
                "from_account": "ACC-001",
                "to_account": "ACC-002",
                "amount": 10.00
            }))
            .to_request();

        let response = test::call_service(&app, request).await;
        let status = response.status();
        let body: serde_json::Value = test::read_body_json(response).await;

        assert_eq!(status, StatusCode::BAD_REQUEST);
        assert_eq!(body["code"], "INVALID_REFERENCE");
    }
}
