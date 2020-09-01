using System;
using System.Collections.Generic;
using System.Text;

namespace DotNet.Exceptions {
    /// <summary>
    /// Typically "429 Too many requests" response codes from an API.
    /// </summary>
    public class ThrottledException : TransientException {
    }
}
