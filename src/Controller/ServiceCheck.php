<?php
/**
 * Statusengine UI
 * Copyright (C) 2016-2017  Daniel Ziegler

 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

namespace Statusengine\Controller;


use Statusengine\Loader\ServiceCheckLoaderInterface;
use Statusengine\ValueObjects\ServiceCheckQueryOptions;

class ServiceCheck extends Controller {

    /**
     * @var ServiceCheckLoaderInterface
     */
    private $ServiceCheckLoader;

    /**
     * ServiceCheck constructor.
     * @param ServiceCheckLoaderInterface $ServiceCheckLoader
     */
    public function __construct(ServiceCheckLoaderInterface $ServiceCheckLoader) {
        $this->ServiceCheckLoader = $ServiceCheckLoader;
    }

    /**
     * @param ServiceCheckQueryOptions $queryOptions
     * @return array
     */
    public function index(ServiceCheckQueryOptions $queryOptions) {
        return $this->ServiceCheckLoader->getServicechecks($queryOptions);
    }


}