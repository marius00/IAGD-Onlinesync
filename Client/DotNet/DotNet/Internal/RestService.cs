using System;
using System.Collections.Generic;
using System.Net.Http;
using System.Net.Http.Headers;
using System.Text;

namespace DotNet.Internal {
    class RestService {
        private readonly HttpClient _client;
        private readonly IJsonSerializer _serializer;

        public RestService(IJsonSerializer serializer, string authToken) {
            _serializer = serializer;

            _client = new HttpClient();
            _client.DefaultRequestHeaders.TryAddWithoutValidation("Simple-Auth", authToken);
            _client.DefaultRequestHeaders.Accept.Add(new MediaTypeWithQualityHeaderValue("application/json"));
        }

        public T Get<T>(string url) {
            var result = _client.GetStringAsync(url).Result;
            return _serializer.DeserializeObject<T>(result);
        }

        public bool Post(string url, string json) {
            var result = _client.PostAsync(url, new StringContent(json, Encoding.UTF8, "application/json")).Result;
            return result.IsSuccessStatusCode;
        }
    }

    // TODO: Improve this, maybe just in the example application?
    /// <summary>
    /// A wrapper for any JSON serializer.
    /// 
    /// Example for newtonsoft serializer:
    /// 
    /// 
    /// private readonly JsonSerializerSettings _settings = new JsonSerializerSettings {
    /// ReferenceLoopHandling = ReferenceLoopHandling.Ignore,
    /// Culture = System.Globalization.CultureInfo.InvariantCulture,
    /// ContractResolver = new Newtonsoft.Json.Serialization.CamelCasePropertyNamesContractResolver()
    /// };
    /// 
    /// JsonConvert.DeserializeObject<T>(responseJson, _settings)
    /// </summary>
    interface IJsonSerializer {
        T DeserializeObject<T>(string value);
    }
}
