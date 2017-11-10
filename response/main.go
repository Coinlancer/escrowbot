package response

import (
    "encoding/json"
    "net/http"
    "./errors"
)

func Json(w http.ResponseWriter, data map[string]interface{}) {
    js, err := json.Marshal(data)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.Write(js)
}

func JsonError(w http.ResponseWriter, err errors.JsonErrorCode, description ...string) {
    resp := map[string]interface{}{
        "error": err,
        "description": "",
    }

    if len(description) > 0 {
        resp["description"] = description[0]
    }

    js, _ := json.Marshal(resp)
    w.WriteHeader(http.StatusInternalServerError)
    w.Header().Set("Content-Type", "application/json")
    w.Write(js)
}
